package job

import (
	"context"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"time"
)

const (
	Default = iota
	FeedParser
)

type Daemon struct {
	logger    *logger.Logger
	scheduler quartz.Scheduler
	jobs      []quartz.Job
}

func NewDaemon(logger *logger.Logger, jobs []quartz.Job) *Daemon {
	scheduler := quartz.NewStdScheduler()

	return &Daemon{
		logger:    logger,
		scheduler: scheduler,
		jobs:      jobs,
	}
}

func (d *Daemon) Start(ctx context.Context) {
	d.scheduler.Start(ctx)

	// TODO replace const time
	for _, job := range d.jobs {
		d.scheduler.ScheduleJob(context.Background(), job, quartz.NewSimpleTrigger(time.Second*30))
	}
}
