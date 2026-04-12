package cronfx

import (
	"context"
	"github.com/duongtuttbn/shared-resource/go-kit/cron"
	"github.com/duongtuttbn/shared-resource/go-kit/log"

	"github.com/samber/lo"
	"go.uber.org/fx"
)

func NewModule() fx.Option {
	return fx.Module("cronfx",
		fx.Provide(
			NewScheduler,
		),
		fx.Invoke(func(scheduler *cron.Scheduler, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					scheduler.Start()
					return nil
				},
				OnStop: func(_ context.Context) error {
					scheduler.Stop()
					return nil
				},
			})
		}),
	)
}

type SchedulerParam struct {
	fx.In
	Jobs []Job `group:"scheduled_jobs"`
}

func NewScheduler(p SchedulerParam) *cron.Scheduler {
	scheduler := cron.New(cron.WithSeconds())

	for _, job := range p.Jobs {
		canOverlap := false
		if j, ok := job.(JobCanOverlap); ok {
			canOverlap = j.CanOverlap()
		}
		lo.Must(scheduler.Add(job.Name(), job.Cron(), func(ctx context.Context) {
			if err := job.Run(ctx); err != nil {
				log.Errorf("error when exectue job: %s, err: %+v", job.Name(), err)
			}
		}, canOverlap))
	}

	return scheduler
}

type Job interface {
	Run(ctx context.Context) error
	Cron() string
	Name() string
}

type JobCanOverlap interface {
	CanOverlap() bool
}

// AsJob register a cron job.
func AsJob(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Job)),
		fx.ResultTags(`group:"scheduled_jobs"`),
	)
}

var (
	_ Job           = jobFuncWrapper{}
	_ JobCanOverlap = jobFuncWrapper{}
)

type jobFuncWrapper struct {
	name       string
	cron       string
	run        func(ctx context.Context) error
	canOverlap bool
}

// NewJobFunc create a new Job from a JobFunc.
func NewJobFunc(name string, cron string, cmd func(ctx context.Context) error, canOverlap ...bool) Job {
	canOverlapConfig := false
	if len(canOverlap) > 0 {
		canOverlapConfig = canOverlap[0]
	}
	return jobFuncWrapper{
		run:        cmd,
		cron:       cron,
		name:       name,
		canOverlap: canOverlapConfig,
	}
}

func (j jobFuncWrapper) Run(ctx context.Context) error {
	return j.run(ctx)
}

func (j jobFuncWrapper) Cron() string {
	return j.cron
}

func (j jobFuncWrapper) Name() string {
	return j.name
}

func (j jobFuncWrapper) CanOverlap() bool {
	return j.canOverlap
}
