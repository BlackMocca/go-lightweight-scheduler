package task

import (
	"context"
	"encoding/json"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
)

type Task struct {
	taskbase `json:",inline"`
	fn       executor.Execution `json:"-"`
}

func NewTask(name string, execution executor.Execution) Execution {
	return &Task{
		taskbase: taskbase{
			taskType: constants.TASK_TYPE_BASE_TASK,
			name:     name,
		},
		fn: execution,
	}
}

func (s Task) MarshalJSON() ([]byte, error) {
	type ptr struct {
		TaskType      string `json:"type"`
		Name          string `json:"name"`
		ExecutionName string `json:"execution_name"`
	}
	sh := ptr{
		TaskType:      string(s.taskbase.taskType),
		Name:          s.taskbase.name,
		ExecutionName: s.fn.GetName(),
	}
	return json.Marshal(sh)
}

func (t Task) GetType() constants.TaskType {
	return t.taskType
}

func (t Task) GetName() string {
	return t.name
}

func (t Task) GetExecutionName() string {
	return t.fn.GetName()
}

func (t Task) Call(ctx context.Context) (interface{}, error) {
	return t.fn.Execute(ctx)
}
