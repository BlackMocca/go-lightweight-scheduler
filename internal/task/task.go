package task

import (
	"context"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
)

type Task struct {
	name string
	fn   executor.Execution
}

func NewTask(name string, execution executor.Execution) Execution {
	return &Task{
		name: name,
		fn:   execution,
	}
}

func (t Task) GetName() string {
	return t.name
}

func (t Task) Call(ctx context.Context) error {
	return t.fn.Execute(ctx)
}
