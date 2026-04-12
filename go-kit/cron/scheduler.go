package cron

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"strconv"
	"sync"
	"time"
	"tla-backend/pkg/go-kit/log"
)

type Scheduler struct {
	c             *cron.Cron
	mu            sync.Mutex
	runningCtxMap map[string]context.CancelFunc
}

func New(opts ...cron.Option) *Scheduler {
	return &Scheduler{
		c:             cron.New(opts...),
		runningCtxMap: make(map[string]context.CancelFunc),
	}
}

func (s *Scheduler) Add(name string, cronStr string, fn func(ctx context.Context), overlap ...bool) (cron.EntryID, error) {
	if cronStr == "" {
		return 0, nil
	}

	canOverlap := false
	if len(overlap) > 0 {
		canOverlap = overlap[0]
	}

	var entryID cron.EntryID
	var err error

	defer func() {
		log.Infof("Registered %s cron: %s, entryID: %d", name, cronStr, entryID)
	}()

	if canOverlap {
		entryID, err = s.c.AddFunc(cronStr, func() {
			log.Infof("Starting " + name)
			start := time.Now()
			runID := fmt.Sprintf("%d:%d", entryID, time.Now().UnixNano())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s.mu.Lock()
			s.runningCtxMap[runID] = cancel
			s.mu.Unlock()

			fn(ctx)

			s.mu.Lock()
			delete(s.runningCtxMap, runID)
			s.mu.Unlock()

			log.Infof("%s took %s", name, time.Since(start))
		})
		return entryID, err
	}

	entryID, err = s.c.AddFunc(cronStr, func() {
		runID := strconv.Itoa(int(entryID))
		s.mu.Lock()
		_, isRunning := s.runningCtxMap[runID]
		if isRunning {
			s.mu.Unlock()
			log.Infof("%s already running, skipping", name)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s.runningCtxMap[runID] = cancel
		s.mu.Unlock()

		log.Infof("Starting " + name)
		start := time.Now()
		fn(ctx)

		s.mu.Lock()
		delete(s.runningCtxMap, runID)
		s.mu.Unlock()

		log.Infof("%s took %s", name, time.Since(start))
	})

	return entryID, err
}

func (s *Scheduler) Start() {
	s.c.Start()
}

func (s *Scheduler) Stop() context.Context {
	ctx := s.c.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cancelFunc := range s.runningCtxMap {
		cancelFunc()
	}

	s.runningCtxMap = make(map[string]context.CancelFunc)

	return ctx
}

func (s *Scheduler) Entries() []cron.Entry {
	return s.c.Entries()
}

func (s *Scheduler) Cron() *cron.Cron {
	return s.c
}
