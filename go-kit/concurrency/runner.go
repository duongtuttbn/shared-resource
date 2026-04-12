package concurrency

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"
)

type Runner[T any] struct {
	jobs                  []Job[T]
	maxConcurrentJobs     int
	rateLimitNumberOfJobs int
	rateLimitDuration     time.Duration
	clearJobsAfterRun     bool
}

type Job[T any] func(ctx context.Context, index int) (T, error)

const defaultMaxConcurrentJobs = 1000

func NewRunner[T any]() *Runner[T] {
	return &Runner[T]{
		jobs:                  make([]Job[T], 0),
		maxConcurrentJobs:     defaultMaxConcurrentJobs,
		clearJobsAfterRun:     true,
		rateLimitNumberOfJobs: -1, // Default no rate limit
	}
}

func (r *Runner[T]) SetMaxConcurrentJobs(maxConcurrentJobs int) *Runner[T] {
	if maxConcurrentJobs == 0 {
		maxConcurrentJobs = defaultMaxConcurrentJobs
	}
	r.maxConcurrentJobs = maxConcurrentJobs
	return r
}

func (r *Runner[T]) SetClearJobsAfterRun(clearJobsAfterRun bool) *Runner[T] {
	r.clearJobsAfterRun = clearJobsAfterRun
	return r
}

func (r *Runner[T]) SetRateLimit(numberOfJobs int, duration time.Duration) *Runner[T] {
	r.rateLimitNumberOfJobs = numberOfJobs
	r.rateLimitDuration = duration
	return r
}

func (r *Runner[T]) AddJob(jobs ...Job[T]) *Runner[T] {
	r.jobs = append(r.jobs, jobs...)
	return r
}

func (r *Runner[T]) Run(ctx context.Context) ([]T, []error, error) {
	if len(r.jobs) == 0 {
		return nil, nil, fmt.Errorf("no jobs to run")
	}

	results := make([]T, len(r.jobs))
	errors := make([]error, len(r.jobs))

	maxConcurrentJobs := r.maxConcurrentJobs

	if r.maxConcurrentJobs <= 0 {
		// No limit
		maxConcurrentJobs = len(r.jobs)
	}

	sem := semaphore.NewWeighted(int64(maxConcurrentJobs))

	var rateLimitBuckets chan int

	if r.rateLimitNumberOfJobs > 0 {
		rateLimitBuckets = make(chan int, r.rateLimitNumberOfJobs)
		defer close(rateLimitBuckets)
		jobDone := make(chan bool, 1)
		defer close(jobDone)
		go func() {
			fillBuckets := func() {
				for i := 0; i < r.rateLimitNumberOfJobs; i++ {
					select {
					case rateLimitBuckets <- 1:
					default:
						return
					}
				}
			}

			rateLimitTicker := time.NewTicker(r.rateLimitDuration).C
			fillBuckets()
			for {
				select {
				case <-rateLimitTicker:
					fillBuckets()
				case <-jobDone:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for jobIndex, job := range r.jobs {
		if err := sem.Acquire(ctx, 1); err != nil {
			return results, errors, err
		}
		go func(resultIndex int, job Job[T]) {
			defer sem.Release(1)
			if rateLimitBuckets != nil {
				<-rateLimitBuckets
			}
			select {
			case <-ctx.Done():
				return
			default:
				result, err := job(ctx, resultIndex)
				results[resultIndex] = result
				errors[resultIndex] = err
			}
		}(jobIndex, job)
	}

	if err := sem.Acquire(ctx, int64(maxConcurrentJobs)); err != nil {
		return results, errors, err
	}

	if r.clearJobsAfterRun {
		r.jobs = make([]Job[T], 0)
	}

	return results, errors, nil
}
