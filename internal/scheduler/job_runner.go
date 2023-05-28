package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/gofrs/uuid"
)

type JobRunner struct {
	id               string
	ctx              context.Context
	ji               *JobInstance
	status           constants.JobStatus
	currentTaskIndex int
	exception        error
	executeDatetime  time.Time
	arguments        *sync.Map
	parameter        *sync.Map
}

func newJobRunner(ctx context.Context, ji *JobInstance) *JobRunner {
	uid, _ := uuid.NewV4()
	runner := &JobRunner{
		id:               uid.String(),
		ctx:              ctx,
		ji:               ji,
		status:           constants.JOB_STATUS_RUNNING,
		currentTaskIndex: 0,
		executeDatetime:  time.Now().UTC(),
		arguments:        new(sync.Map),
		parameter:        new(sync.Map),
	}

	ctx = context.WithValue(ctx, constants.JOB_RUNNER_INSTANCE_KEY, runner)
	runner.ctx = ctx
	if ji.arguments != nil {
		for k, v := range ji.arguments {
			runner.arguments.Store(k, v)
		}
	}

	return runner
}

func (jr JobRunner) GetId() string {
	return jr.id
}

func (jr JobRunner) GetStatus() constants.JobStatus {
	return jr.status
}

func (jr JobRunner) GetTask() task.Execution {
	return jr.ji.tasks[jr.currentTaskIndex]
}

func (jr JobRunner) GetException() error {
	return jr.exception
}

func (jr JobRunner) GetExecuteDatetime() time.Time {
	return jr.executeDatetime
}

func (jr JobRunner) GetArguments() *sync.Map {
	return jr.arguments
}

func (jr JobRunner) GetParameter() *sync.Map {
	return jr.parameter
}

func (jr *JobRunner) run() {
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

	for index, task := range jr.ji.tasks {
		jr.currentTaskIndex = index
		if err := task.Call(jr.ctx); err != nil {
			jr.exception = err
			return
		}
	}

	jr.status = constants.JOB_STATUS_SUCCESS
}
