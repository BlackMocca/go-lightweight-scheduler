package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/connection"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/logger"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/go-co-op/gocron"
)

type SchedulerInstance struct {
	Scheduler      *gocron.Scheduler
	cronExpression string
	name           string
	description    string
	jobInstance    *JobInstance
	config         SchedulerConfig
	logger         *logger.Log
	dbAdapter      connection.DatabaseAdapterConnection
}

func NewScheduler(cronExpression string, name string, description string, config SchedulerConfig) *SchedulerInstance {
	scheduler := gocron.NewScheduler(time.Local)
	scheduler.SetMaxConcurrentJobs(config.MaxActiveConcurrent, gocron.WaitMode)

	return &SchedulerInstance{
		Scheduler:      scheduler,
		name:           name,
		description:    description,
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

func (s *SchedulerInstance) SetAdapter(dbAdapter connection.DatabaseAdapterConnection) {
	s.dbAdapter = dbAdapter
}

func (s SchedulerInstance) GetAdapter() connection.DatabaseAdapterConnection {
	return s.dbAdapter
}

func (s *SchedulerInstance) Start() error {
	if s.cronExpression != "" {
		_, fn := s.jobInstance.trigger("", nil, nil)
		job, err := s.Scheduler.Cron(s.cronExpression).Do(fn)
		if err != nil {
			return err
		}
		s.jobInstance.Job = job
	}

	s.Scheduler.StartAsync()
	s.logger.Info("start scheduler", map[string]interface{}{
		"scheduler_cronjob_expression": s.cronExpression,
		"scheduler_name":               s.name,
	})
	return nil
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
	jobInstance.SetScheduler(s)
	s.jobInstance = jobInstance

	return nil
}

func (s *SchedulerInstance) Run(trigger *models.Trigger) string {
	/* ตั้งเวลาล่วงหน้า */
	if trigger.ExecuteDatetime != (time.Time{}) && trigger.ExecuteDatetime.Sub(time.Now()) > 0 {
		duration := trigger.ExecuteDatetime.Sub(time.Now())
		jobId, fn := s.jobInstance.trigger(trigger.JobId, trigger.GetConfigMutex(), &trigger.ExecuteDatetime)

		trigger.JobId = jobId
		go func(trigger models.Trigger, duration time.Duration, call func()) {
			time.Sleep(duration)
			trigger.IsTrigger = true
			checkTrigger, _ := s.dbAdapter.GetRepository().GetOneTriggerByJobId(context.Background(), trigger.JobId)
			if checkTrigger != nil && checkTrigger.IsActive {
				fn()
			}
			s.dbAdapter.GetRepository().UpsertTrigger(context.Background(), &trigger)
		}(*trigger, duration, fn)
		return jobId
	}
	/* run ทันที */
	jobId, fn := s.jobInstance.trigger(trigger.JobId, trigger.GetConfigMutex(), nil)
	go fn()
	trigger.JobId = jobId
	trigger.IsTrigger = true
	go s.dbAdapter.GetRepository().UpsertTrigger(context.Background(), trigger)
	return jobId
}
