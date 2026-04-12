package cron_test

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
	"tla-backend/pkg/go-kit/cron"
	"tla-backend/pkg/go-kit/log"

	"github.com/samber/lo"
)

func TestSchedulerShutdown(_ *testing.T) {
	log.SetDefault(log.Config{})
	scheduler := cron.New(cron.WithSeconds())

	lo.Must(scheduler.Add("tick1", "* * * * * *", func(_ context.Context) {
		println("tick1")
	}))

	lo.Must(scheduler.Add("tick", "* * * * * *", func(ctx context.Context) {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				println("tick")
			}
		}
	}))

	if len(scheduler.Entries()) == 0 {
		log.Info("No cron configured")
		return
	}

	osSignal := make(chan os.Signal, 1)

	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)

	scheduler.Start()

	defer func() {
		log.Info("Shutting down scheduler")
		ctx := scheduler.Stop()
		<-ctx.Done()
	}()

	s := <-osSignal
	println("Got signal:", s.String())
}
