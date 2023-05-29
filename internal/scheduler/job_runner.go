package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/gofrs/uuid"
)

type JobRunner interface {
	GetId() string
	GetStatus() constants.JobStatus
	GetTask() task.Execution
	GetException() error
	GetExecuteDatetime() time.Time
	GetArguments() *sync.Map
	GetParameter() *sync.Map
	GetTaskValue(taskName string) (data interface{}, ok bool)
}

type jobRunner struct {
	id                  string
	ctx                 context.Context
	tasks               []task.Execution
	status              constants.JobStatus
	currentTaskIndex    int
	exceptionOnTaskName string
	exception           error
	executeDatetime     time.Time
	arguments           *sync.Map
	parameter           *sync.Map
	taskValue           *sync.Map
}

func newJobRunner(ctx context.Context, ji *JobInstance) *jobRunner {
	uid, _ := uuid.NewV4()
	runner := &jobRunner{
		id:               uid.String(),
		ctx:              ctx,
		tasks:            ji.tasks,
		status:           constants.JOB_STATUS_RUNNING,
		currentTaskIndex: 0,
		executeDatetime:  time.Now().UTC(),
		arguments:        new(sync.Map),
		parameter:        new(sync.Map),
		taskValue:        new(sync.Map),
	}

	if ji.arguments != nil {
		for k, v := range ji.arguments {
			runner.arguments.Store(k, v)
		}
	}

	return runner
}

func (jr *jobRunner) getRunnerInterface() JobRunner {
	return jr
}

func (jr jobRunner) GetId() string {
	return jr.id
}

func (jr jobRunner) GetStatus() constants.JobStatus {
	return jr.status
}

func (jr jobRunner) GetTask() task.Execution {
	return jr.tasks[jr.currentTaskIndex]
}

func (jr jobRunner) GetException() error {
	return jr.exception
}

func (jr jobRunner) GetExecuteDatetime() time.Time {
	return jr.executeDatetime
}

func (jr jobRunner) GetArguments() *sync.Map {
	return jr.arguments
}

func (jr jobRunner) GetParameter() *sync.Map {
	return jr.parameter
}

func (jr jobRunner) GetTaskValue(taskName string) (data interface{}, ok bool) {
	return jr.taskValue.Load(taskName)
}

func (jr *jobRunner) run(tasks []task.Execution) {
	defer func() {
		if r := recover(); r != nil {
			jr.exception = r.(error)

		}
	}()
	defer func() {
		if jr.exception != nil {
			jr.status = constants.JOB_STATUS_ERROR
		}
	}()

	jr.tasks = tasks
	jr.currentTaskIndex = 0
	for index, taskExecution := range tasks {
		jr.currentTaskIndex = index
		value, err := taskExecution.Call(jr.ctx)
		if err != nil {
			jr.exceptionOnTaskName = taskExecution.GetName()
			jr.exception = err
			return
		}
		jr.taskValue.Store(taskExecution.GetName(), value)

		switch taskExecution.GetType() {
		case constants.TASK_TYPE_BRANCH_TASK:
			val := value.(task.TaskBranchPipeLine)
			tasks := val.GetTasks()
			jr.run(tasks)
			return
		}
	}

	jr.status = constants.JOB_STATUS_SUCCESS
}

func (jr *jobRunner) clear() {
	jr.arguments = nil
	jr.parameter = nil
	jr.exception = nil
	jr.taskValue = nil
	jr.tasks = nil
}
