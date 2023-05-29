package task

import (
	"context"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
)

type Execution interface {
	GetType() constants.TaskType
	GetName() string
	Call(ctx context.Context) (interface{}, error)
}

type taskbase struct {
	taskType constants.TaskType
	name     string
}
