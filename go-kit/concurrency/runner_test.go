package concurrency

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func setup(numberOfJobs int, maxConcurrentJobs int, jobSleep int, rateLimit int, duration time.Duration) *Runner[string] {
	runner := NewRunner[string]().
		SetMaxConcurrentJobs(maxConcurrentJobs).
		SetRateLimit(rateLimit, duration)

	for i := 0; i < numberOfJobs; i++ {
		runner.AddJob(func(_ context.Context, index int) (string, error) {
			if jobSleep > 0 {
				time.Sleep(time.Duration(jobSleep) * time.Millisecond)
			}
			return fmt.Sprintf("%d", index), nil
		})
	}
	return runner
}

func TestGoRoutineRunner(t *testing.T) {
	runner := setup(12, -1, 0, 2, 10*time.Millisecond)
	start := time.Now()
	results, jobErrors, err := runner.Run(context.Background())
	elapsed := time.Since(start)
	require.NoError(t, err)
	for _, err2 := range jobErrors {
		require.NoError(t, err2)
	}

	require.Greater(t, 55*time.Millisecond, elapsed)
	require.Less(t, 50*time.Millisecond, elapsed)

	require.Len(t, results, 12)
	require.Equal(t, "0", results[0])
	require.Equal(t, "9", results[9])
}

func TestGoRoutineRunnerWithDeadlineContext(t *testing.T) {
	runner := setup(12, 0, 0, 2, 10*time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(31*time.Millisecond))
	defer cancel()

	start := time.Now()
	results, jobErrors, err := runner.Run(ctx)
	elapsed := time.Since(start)
	require.Error(t, err)
	for _, err2 := range jobErrors {
		require.NoError(t, err2)
	}

	require.Greater(t, 35*time.Millisecond, elapsed)
	require.Less(t, 30*time.Millisecond, elapsed)

	require.Len(t, results, 12)
	notEmptyResults := 0
	for _, result := range results {
		if result != "" {
			notEmptyResults++
		}
	}
	require.Equal(t, 8, notEmptyResults)
}

func TestGoRoutineRunnerWithCancelContext(t *testing.T) {
	runner := setup(12, 2, 0, 2, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	results, jobErrors, err := runner.Run(ctx)
	elapsed := time.Since(start)
	require.Error(t, err)
	for _, err2 := range jobErrors {
		require.NoError(t, err2)
	}

	require.Greater(t, 27*time.Millisecond, elapsed)
	require.Less(t, 25*time.Millisecond, elapsed)

	require.Len(t, results, 12)
	notEmptyResults := 0
	for _, result := range results {
		if result != "" {
			notEmptyResults++
		}
	}
	require.Equal(t, 6, notEmptyResults)
}

func TestGoRoutineRunnerWithSlowJob(t *testing.T) {
	runner := setup(12, 2, 15, 2, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	results, jobErrors, err := runner.Run(ctx)
	elapsed := time.Since(start)
	require.Error(t, err)
	for _, err2 := range jobErrors {
		require.NoError(t, err2)
	}

	require.Greater(t, 27*time.Millisecond, elapsed)
	require.Less(t, 25*time.Millisecond, elapsed)

	require.Len(t, results, 12)
	notEmptyResults := 0
	for _, result := range results {
		if result != "" {
			notEmptyResults++
		}
	}
	require.Equal(t, 2, notEmptyResults)
}
