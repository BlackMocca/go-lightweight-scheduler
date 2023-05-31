package task

import (
	"context"
	"encoding/json"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
)

type Execution interface {
	GetType() constants.TaskType
	GetName() string
	GetExecutionName() string
	Call(ctx context.Context) (interface{}, error)
	MarshalJSON() ([]byte, error)
}

type taskbase struct {
	taskType constants.TaskType
	name     string
}

func (s taskbase) MarshalJSON() ([]byte, error) {
	type ptr struct {
		TaskType string `json:"type"`
		Name     string `json:"name"`
	}
	sh := ptr{
		TaskType: string(s.taskType),
		Name:     s.name,
	}
	return json.Marshal(sh)
}
