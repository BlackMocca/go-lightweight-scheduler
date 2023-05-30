package scheduler

import (
	"encoding/json"
	"errors"
	"sync"
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
		Name        string                   `json:"name"`
		Cronjob     string                   `json:"cronjob_expression"`
		IsRunning   bool                     `json:"is_running"`
		Arguments   map[string]interface{}   `json:"arguments"`
		LastRun     string                   `json:"last_run"`
		NextRun     string                   `json:"next_run"`
		PreviousRun string                   `json:"previous_run"`
		Config      SchedulerConfig          `json:"config"`
		Tasks       []map[string]interface{} `json:"tasks"`
	}
	var sh = ptr{
		Name:      s.name,
		Cronjob:   s.cronExpression,
		IsRunning: s.Scheduler.IsRunning(),
		Config:    s.config,
		Tasks:     make([]map[string]interface{}, 0),
	}
	if s.jobInstance != nil {
		sh.Arguments = s.jobInstance.arguments
		if s.jobInstance.Job != nil {
			sh.NextRun = s.jobInstance.Job.NextRun().Format(constants.TIME_FORMAT_RFC339)
			sh.LastRun = s.jobInstance.Job.LastRun().Format(constants.TIME_FORMAT_RFC339)
			sh.PreviousRun = s.jobInstance.Job.PreviousRun().Format(constants.TIME_FORMAT_RFC339)
		}
		if len(s.jobInstance.tasks) > 0 {
			for _, task := range s.jobInstance.tasks {
				bu, _ := task.MarshalJSON()
				m := map[string]interface{}{}
				json.Unmarshal(bu, &m)

				sh.Tasks = append(sh.Tasks, m)
			}
		}
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
	jobInstance.SetScheduler(s.name, s.config)

	_, fn := jobInstance.trigger(nil, nil)
	job, err := s.Scheduler.Cron(s.cronExpression).Do(fn)
	if err != nil {
		return err
	}

	jobInstance.Job = job
	s.jobInstance = jobInstance
	return nil
}

func (s *SchedulerInstance) Run(triggerConfig *sync.Map, triggerTime time.Time) string {
	/* ตั้งเวลาล่วงหน้า */
	if triggerTime != (time.Time{}) && triggerTime.Sub(time.Now()) > 0 {
		duration := triggerTime.Sub(time.Now())
		jobId, fn := s.jobInstance.trigger(triggerConfig, &triggerTime)

		go func(duration time.Duration, call func()) {
			time.Sleep(duration)
			fn()
		}(duration, fn)
		return jobId
	}
	/* run ทันที */
	jobId, fn := s.jobInstance.trigger(triggerConfig, nil)
	go fn()
	return jobId
}
