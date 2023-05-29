package task

import (
	"context"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
)

type Task struct {
	taskbase
	fn executor.Execution
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

func (t Task) GetType() constants.TaskType {
	return t.taskType
}

func (t Task) GetName() string {
	return t.name
}

func (t Task) Call(ctx context.Context) (interface{}, error) {
	return t.fn.Execute(ctx)
}
