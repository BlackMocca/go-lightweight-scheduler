package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
)

const (
	default_max_active_concurrent = 32
)

var (
	default_func = func(ctx context.Context) error {
		return nil
	}
)

type SchedulerConfig struct {
	MaxActiveConcurrent int
	RetryTimes          int
	RetryDelay          time.Duration
	JobTimeout          time.Duration
	JobMode             constants.JobMode
	OnSuccess           func(ctx context.Context) error
	OnError             func(ctx context.Context) error
}

func NewDefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		MaxActiveConcurrent: default_max_active_concurrent,
		RetryTimes:          0,
		RetryDelay:          0,
		JobMode:             constants.JOB_MODE_CONCURRENT,
		OnSuccess:           nil,
		OnError:             nil,
	}
}

func (s SchedulerConfig) MarshalJSON() ([]byte, error) {
	type ptr struct {
		MaxActiveConcurrent int  `json:"max_active_concurrent"`
		RetryTimes          int  `json:"retry_times"`
		RetryDelay          int  `json:"retry_delay"`
		JobTimeout          int  `json:"job_timeout"`
		JobMode             int8 `json:"job_mode"`
		OnSuccess           bool `json:"is_handle_on_success"`
		OnError             bool `json:"is_handle_on_error"`
	}
	var sh = ptr{
		MaxActiveConcurrent: s.MaxActiveConcurrent,
		RetryTimes:          s.RetryTimes,
		RetryDelay:          int(s.RetryDelay),
		JobTimeout:          int(s.JobTimeout),
		JobMode:             int8(s.JobMode),
		OnSuccess:           s.OnSuccess != nil,
		OnError:             s.OnError != nil,
	}

	return json.Marshal(sh)
}
