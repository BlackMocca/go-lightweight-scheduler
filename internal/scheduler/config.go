package scheduler

import (
	"context"
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
		OnSuccess:           default_func,
		OnError:             default_func,
	}
}
