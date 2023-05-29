package task

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/spf13/cast"
)

type TaskBranch struct {
	taskbase
	fn          executor.Execution
	taskBranchs []TaskBranchPipeLine
}

func NewTaskBranch(name string, execution executor.Execution, tasks []TaskBranchPipeLine) Execution {
	return &TaskBranch{
		taskbase: taskbase{
			taskType: constants.TASK_TYPE_BRANCH_TASK,
			name:     name,
		},
		fn:          execution,
		taskBranchs: tasks,
	}
}

func (t TaskBranch) GetType() constants.TaskType {
	return t.taskType
}

func (t TaskBranch) GetName() string {
	return t.name
}

func (t TaskBranch) Call(ctx context.Context) (interface{}, error) {
	taskname, err := t.fn.Execute(ctx)
	if err != nil {
		return nil, err
	}
	if reflect.TypeOf(taskname).Kind() != reflect.String {
		return nil, errors.New("TaskBranch must be return taskanme on type string")
	}
	for _, task := range t.taskBranchs {
		if cast.ToString(taskname) == task.name {
			return task, nil
		}
	}
	return nil, fmt.Errorf("Task %s not found in TaskBranch", cast.ToString(taskname))
}

type TaskBranchPipeLine struct {
	name  string
	tasks []Execution
}

func NewTaskBranchPipeline(pipes map[string][]Execution) []TaskBranchPipeLine {
	var ptrs = make([]TaskBranchPipeLine, 0)
	for name, tasks := range pipes {
		ptrs = append(ptrs, TaskBranchPipeLine{
			name:  name,
			tasks: tasks,
		})
	}
	return ptrs
}

func (t TaskBranchPipeLine) GetName() string {
	return t.name
}

func (t TaskBranchPipeLine) GetTasks() []Execution {
	return t.tasks
}
