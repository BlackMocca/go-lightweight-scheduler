package scheduler

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/logger"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/gofrs/uuid"
)

type JobRunner interface {
	GetId() string
	GetSchedulerName() string
	GetStatus() constants.JobStatus
	GetTask() task.Execution
	GetException() error
	GetExecuteDatetime() time.Time
	GetArguments() *sync.Map     // static data when job run
	GetParameter() *sync.Map     // pass data through pipeline
	GetTriggerConfig() *sync.Map // pass data when trigger
	GetTaskValue(taskName string) (data interface{}, ok bool)
	GetLogger() *logger.Log
}

type Exception interface {
	Error() string
	StackTrace() string
}

type jobRunner struct {
	id                  string
	schedulerName       string
	ctx                 context.Context
	tasks               []task.Execution
	status              constants.JobStatus
	currentTaskIndex    int
	exceptionOnTaskName string
	exception           Exception
	executeDatetime     time.Time
	arguments           *sync.Map
	parameter           *sync.Map
	taskValue           *sync.Map
	logger              *logger.Log
	taskResults         []taskResult
	triggerConfig       *sync.Map
}

type taskResult struct {
	task        task.Execution
	status      constants.JobStatus
	startDate   time.Time
	endDatetime time.Time
}

type runnerException struct {
	err   error
	stack []byte
}

func newRunnerException(err error) Exception {
	return &runnerException{
		err:   err,
		stack: debug.Stack(),
	}
}

func (r *runnerException) Error() string {
	return r.err.Error()
}
func (r *runnerException) StackTrace() string {
	return string(r.stack)
}

func newJobRunner(ctx context.Context, ji *JobInstance, triggerConfig *sync.Map, executeDatetime *time.Time) *jobRunner {
	uid, _ := uuid.NewV4()
	runner := &jobRunner{
		id:               uid.String(),
		schedulerName:    ji.schedulerName,
		ctx:              ctx,
		tasks:            ji.tasks,
		status:           constants.JOB_STATUS_WAITING,
		currentTaskIndex: 0,
		executeDatetime:  time.Now(),
		arguments:        new(sync.Map),
		parameter:        new(sync.Map),
		taskValue:        new(sync.Map),
		triggerConfig:    new(sync.Map),
	}

	if ji.arguments != nil {
		for k, v := range ji.arguments {
			runner.arguments.Store(k, v)
		}
	}
	if triggerConfig != nil {
		runner.triggerConfig = triggerConfig
	}
	if executeDatetime != nil {
		runner.executeDatetime = *executeDatetime
	}

	return runner
}

func (jr *jobRunner) getRunnerInterface() JobRunner {
	return jr
}

func (jr jobRunner) GetId() string {
	return jr.id
}

func (jr jobRunner) GetSchedulerName() string {
	return jr.schedulerName
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

func (jr jobRunner) GetTriggerConfig() *sync.Map {
	return jr.triggerConfig
}

func (jr jobRunner) GetTaskValue(taskName string) (data interface{}, ok bool) {
	return jr.taskValue.Load(taskName)
}

func (jr jobRunner) GetLogger() *logger.Log {
	return jr.logger
}

func (jr *jobRunner) run(tasks []task.Execution) {
	defer func() {
		if r := recover(); r != nil {
			jr.exception = newRunnerException(r.(error))
		}
	}()
	defer func() {
		if jr.exception != nil {
			jr.status = constants.JOB_STATUS_ERROR
		}
	}()

	jr.status = constants.JOB_STATUS_RUNNING
	jr.tasks = tasks
	jr.currentTaskIndex = 0
	jr.taskResults = make([]taskResult, len(jr.tasks))
	for index, taskExecution := range tasks {
		pathfile := constants.LOG_PATH_RUNNER_TASK(jr.schedulerName, jr.executeDatetime, taskExecution.GetName())
		jr.logger = logger.NewLoggerWithFile(pathfile)

		taskResult := taskResult{
			status:      constants.JOB_STATUS_RUNNING,
			startDate:   time.Now(),
			endDatetime: time.Now(),
		}
		jr.logger.Info(fmt.Sprintf("scheduler %s with starting task %s", jr.schedulerName, taskExecution.GetName()), map[string]interface{}{"jobId": jr.id, "task_name": taskExecution.GetName(), "scheduler_name": jr.schedulerName, "task_start": taskResult.startDate.Format(constants.TIME_FORMAT_RFC339)})

		jr.currentTaskIndex = index
		value, err := taskExecution.Call(jr.ctx)
		taskResult.endDatetime = time.Now()
		if err != nil {
			jr.exceptionOnTaskName = taskExecution.GetName()
			jr.exception = newRunnerException(err)
			taskResult.status = constants.JOB_STATUS_ERROR
			jr.taskResults[index] = taskResult
			jr.logger.Error(err, map[string]interface{}{"jobId": jr.id, "task_name": taskExecution.GetName(), "scheduler_name": jr.schedulerName, "task_start": taskResult.startDate.Format(constants.TIME_FORMAT_RFC339), "task_end": taskResult.endDatetime.Format(constants.TIME_FORMAT_RFC339), "task_status": taskResult.status})
			return
		}
		taskResult.status = constants.JOB_STATUS_SUCCESS
		jr.taskResults[index] = taskResult
		jr.taskValue.Store(taskExecution.GetName(), value)

		jr.logger.Info(fmt.Sprintf("scheduler %s with ending task %s", jr.schedulerName, taskExecution.GetName()), map[string]interface{}{"jobId": jr.id, "task_name": taskExecution.GetName(), "scheduler_name": jr.schedulerName, "task_start": taskResult.startDate.Format(constants.TIME_FORMAT_RFC339), "task_end": taskResult.endDatetime.Format(constants.TIME_FORMAT_RFC339), "task_status": taskResult.status})

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
	jr = nil
}
