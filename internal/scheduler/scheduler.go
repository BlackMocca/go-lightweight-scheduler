package scheduler

import (
	"errors"
	"time"

	"github.com/go-co-op/gocron"
)

type SchedulerInstance struct {
	Scheduler      *gocron.Scheduler
	cronExpression string
	name           string
	jobInstance    *JobInstance
	config         SchedulerConfig
}

func NewScheduler(cronExpression string, name string, config SchedulerConfig) *SchedulerInstance {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.SetMaxConcurrentJobs(config.MaxActiveConcurrent, gocron.WaitMode)

	return &SchedulerInstance{
		Scheduler:      scheduler,
		name:           name,
		cronExpression: cronExpression,
		config:         config,
	}
}

func (s SchedulerInstance) Start() {
	s.Scheduler.StartAsync()
}

func (s SchedulerInstance) Stop() {
	s.Scheduler.Stop()
}

func (s *SchedulerInstance) RegisterJob(jobInstance *JobInstance) error {
	if jobInstance.GetTotalTask() == 0 {
		return errors.New("required any task in jobInstance")
	}

	job, err := s.Scheduler.Every(1).Second().Do(jobInstance.onBefore)
	// job, err := s.Scheduler.Cron(s.cronExpression).Do(jobInstance.onBefore)
	if err != nil {
		return err
	}
	jobInstance.Job = job
	jobInstance.SetSchedulerConfig(s.config)

	s.jobInstance = jobInstance
	s.Scheduler.StartAsync()
	return nil
}
