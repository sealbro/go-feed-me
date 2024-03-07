package job

import (
	"context"
	"fmt"
	"github.com/reugn/go-quartz/quartz"
	"github.com/sealbro/go-feed-me/pkg/logger"
)

const (
	FeedParser = iota
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

func (d *Daemon) Start(ctx context.Context) error {
	d.scheduler.Start(ctx)

	trigger, err := quartz.NewCronTrigger(d.config.Cron)
	if err != nil {
		return fmt.Errorf("can't create cron trigger: %w", err)
	}

	// TODO replace const time
	for _, job := range d.jobs {
		jobDetail := quartz.NewJobDetail(job, quartz.NewJobKey(job.Description()))
		err := d.scheduler.ScheduleJob(jobDetail, trigger)
		if err != nil {
			return fmt.Errorf("can't schedule job: %w", err)
		}
	}

	return nil
}
