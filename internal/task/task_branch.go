package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/spf13/cast"
)

type TaskBranch struct {
	taskbase    `json:",inline"`
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

func (s TaskBranch) MarshalJSON() ([]byte, error) {
	type ptr struct {
		TaskType      string              `json:"type"`
		Name          string              `json:"name"`
		ExecutionName string              `json:"execution_name"`
		TaskBranchs   TaskBranchPipeLines `json:"task_branchs"`
	}
	sh := ptr{
		TaskType:      string(s.taskbase.taskType),
		Name:          string(s.taskbase.name),
		ExecutionName: s.fn.GetName(),
		TaskBranchs:   s.taskBranchs,
	}
	return json.Marshal(sh)
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

type TaskBranchPipeLines []TaskBranchPipeLine

func (s TaskBranchPipeLines) MarshalJSON() ([]byte, error) {
	var ptr = map[string][]map[string]interface{}{}
	for _, item := range s {
		if _, ok := ptr[item.name]; !ok {
			ptr[item.name] = make([]map[string]interface{}, 0)
		}
		if item.tasks != nil && len(item.tasks) > 0 {
			for _, task := range item.tasks {
				bu, _ := task.MarshalJSON()
				m := map[string]interface{}{}
				json.Unmarshal(bu, &m)

				ptr[item.name] = append(ptr[item.name], m)
			}
		}
	}
	return json.Marshal(ptr)
}

func (s *TaskBranchPipeLine) MarshalJSON() ([]byte, error) {
	type ptr struct {
		Name  string                   `json:"name"`
		Tasks []map[string]interface{} `json:"tasks"`
	}
	var sh = ptr{
		Name:  s.name,
		Tasks: make([]map[string]interface{}, 0),
	}
	if s.tasks != nil && len(s.tasks) > 0 {
		for _, task := range s.tasks {
			bu, _ := task.MarshalJSON()
			m := map[string]interface{}{}
			json.Unmarshal(bu, &m)

			sh.Tasks = append(sh.Tasks, m)
		}
	}
	return json.Marshal(sh)
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
