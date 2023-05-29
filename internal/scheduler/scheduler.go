package scheduler

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/logger"
	"github.com/go-co-op/gocron"
)

type SchedulerInstance struct {
	Scheduler      *gocron.Scheduler
	cronExpression string
	name           string
	jobInstance    *JobInstance
	config         SchedulerConfig
	logger         *logger.Log
}

func NewScheduler(cronExpression string, name string, config SchedulerConfig) *SchedulerInstance {
	scheduler := gocron.NewScheduler(time.Local)
	scheduler.SetMaxConcurrentJobs(config.MaxActiveConcurrent, gocron.WaitMode)

	return &SchedulerInstance{
		Scheduler:      scheduler,
		name:           name,
		cronExpression: cronExpression,
		config:         config,
		logger:         logger.NewLoggerWithFile(constants.LOG_PATH_SCHEDULER),
	}
}

func (s *SchedulerInstance) MarshalJSON() ([]byte, error) {
	type ptr struct {
		Name      string                 `json:"name"`
		Cronjob   string                 `json:"cronjob_expression"`
		IsRunning bool                   `json:"is_running"`
		Arguments map[string]interface{} `json:"arguments"`
		Config    SchedulerConfig        `json:"config"`
	}
	var sh = ptr{
		Name:      s.name,
		Cronjob:   s.cronExpression,
		IsRunning: s.Scheduler.IsRunning(),
		Config:    s.config,
	}
	if s.jobInstance != nil {
		sh.Arguments = s.jobInstance.arguments
	}
	return json.Marshal(sh)
}

func (s SchedulerInstance) GetName() string {
	return s.name
}

func (s SchedulerInstance) GetCronjobExpression() string {
	return s.cronExpression
}

func (s SchedulerInstance) Start() {
	s.Scheduler.StartAsync()
	s.logger.Info("start scheduler", map[string]interface{}{
		"scheduler_cronjob_expression": s.cronExpression,
		"scheduler_name":               s.name,
	})
}

func (s SchedulerInstance) Stop() {
	s.Scheduler.Stop()
	s.logger.Info("stop scheduler", map[string]interface{}{
		"scheduler_name": s.name,
	})
}

func (s *SchedulerInstance) RegisterJob(jobInstance *JobInstance) error {
	if jobInstance.GetTotalTask() == 0 {
		return errors.New("required any task in jobInstance")
	}

	// job, err := s.Scheduler.Every(1).Second().Do(jobInstance.onBefore)
	job, err := s.Scheduler.Cron(s.cronExpression).Do(jobInstance.onBefore)
	if err != nil {
		return err
	}
	jobInstance.Job = job
	jobInstance.SetScheduler(s.name, s.config)

	s.jobInstance = jobInstance
	return nil
}
