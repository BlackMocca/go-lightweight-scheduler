package models

import (
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
)

type Job struct {
	TableName       struct{}            `json:"-" db:"jobs"`
	SchedulerName   string              `json:"scheduler_name" db:"scheduler_name"`
	JobId           string              `json:"job_id" db:"job_id"`
	Status          constants.JobStatus `json:"status" db:"status"`
	StartDateTime   *time.Time          `json:"start_datetime" db:"start_datetime"`
	EndDatetime     *time.Time          `json:"end_datetime" db:"end_datetime"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" db:"updated_at"`
	Trigger         *Trigger            `json:"trigger" db:"-"`
	JobRunningTasks []*JobTask          `json:"job_running_tasks" db:"-"`
}

type JobTask struct {
	TableName     struct{}            `json:"-" db:"job_taks"`
	Id            int64               `json:"id" db:"id"`
	SchedulerName string              `json:"scheduler_name" db:"scheduler_name"`
	JobId         string              `json:"job_id" db:"job_id"`
	Status        constants.JobStatus `json:"status" db:"task_status"`
	TaskName      string              `json:"name" db:"task_name"`
	TaskType      string              `json:"type" db:"task_type"`
	ExecutionName string              `json:"execution_name" db:"execution_name"`
	StartDateTime time.Time           `json:"start_datetime" db:"start_datetime"`
	EndDatetime   *time.Time          `json:"end_datetime" db:"end_datetime"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" db:"updated_at"`
	TaskException string              `json:"exception" db:"exception"`
	StackTrace    string              `json:"stacktrace" db:"stacktrace"`
}
