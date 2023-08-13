package job

import (
	"context"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/pkg/logger"
)

const (
	Default = iota
	FeedParser
)

type Daemon struct {
	logger    *logger.Logger
	scheduler quartz.Scheduler
	jobs      []quartz.Job
	config    *DaemonConfig
}

func NewDaemon(logger *logger.Logger, config *DaemonConfig, jobs []quartz.Job) *Daemon {
	scheduler := quartz.NewStdScheduler()

	return &Daemon{
		logger:    logger,
		scheduler: scheduler,
		jobs:      jobs,
		config:    config,
	}
}

func (d *Daemon) Start(ctx context.Context) {
	d.scheduler.Start(ctx)

	trigger, err := quartz.NewCronTrigger(d.config.Cron)
	if err != nil {
		d.logger.Sugar().Ctx(ctx).Fatalf("can't create cron trigger: %v", err)
	}

	// TODO replace const time
	for _, job := range d.jobs {
		err := d.scheduler.ScheduleJob(ctx, job, trigger)
		if err != nil {
			d.logger.Sugar().Ctx(ctx).Fatalf("can't schedule job: %v", err)
		}
	}
}
