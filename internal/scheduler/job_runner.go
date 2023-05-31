package scheduler

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/connection"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/logger"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
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
	endDatetime         *time.Time
	arguments           *sync.Map
	parameter           *sync.Map
	taskValue           *sync.Map
	logger              *logger.Log
	taskResults         []taskResult
	triggerConfig       *sync.Map
	dbAdapter           connection.DatabaseAdapterConnection
	triggerType         constants.TriggerType
	logjob              *models.Job     // for save in db
	logtaskrunning      *models.JobTask // for save in db
}

type taskResult struct {
	task        task.Execution
	status      constants.JobStatus
	startDate   time.Time
	endDatetime *time.Time
}

type runnerException struct {
	err   error
	stack []byte
}

func newRunnerException(err error, withStackTrace bool) Exception {
	ex := &runnerException{
		err: err,
	}
	if withStackTrace {
		ex.stack = debug.Stack()
	}
	return ex
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
		schedulerName:    ji.scheduler.name,
		ctx:              ctx,
		tasks:            ji.tasks,
		status:           constants.JOB_STATUS_WAITING,
		currentTaskIndex: 0,
		executeDatetime:  time.Now(),
		arguments:        new(sync.Map),
		parameter:        new(sync.Map),
		taskValue:        new(sync.Map),
		triggerConfig:    new(sync.Map),
		dbAdapter:        ji.scheduler.dbAdapter,
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
			if reflect.TypeOf(r).Kind() == reflect.String {
				jr.exception = newRunnerException(errors.New(r.(string)), true)
				jr.setStatus(constants.JOB_STATUS_FAILED)
			} else {
				jr.exception = newRunnerException(r.(error), true)
				jr.setStatus(constants.JOB_STATUS_FAILED)
			}
		}
	}()

	/* save processing on task but without error */
	saveJobTask := func(runner *jobRunner, taskExecution task.Execution, taskResult taskResult) error {
		jobtask := &models.JobTask{
			JobId:         jr.id,
			SchedulerName: jr.schedulerName,
			Status:        taskResult.status,
			TaskName:      taskExecution.GetName(),
			TaskType:      string(taskExecution.GetType()),
			ExecutionName: taskExecution.GetExecutionName(),
			StartDateTime: taskResult.startDate,
			EndDatetime:   taskResult.endDatetime,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if runner.exception != nil {
			jobtask.TaskException = runner.exception.Error()
			jobtask.StackTrace = runner.exception.StackTrace()
		}
		jr.logtaskrunning = jobtask
		return jr.dbAdapter.GetRepository().UpsertJobTask(jr.ctx, jobtask)
	}

	jr.setStatus(constants.JOB_STATUS_RUNNING)
	jr.tasks = tasks
	jr.currentTaskIndex = 0
	jr.taskResults = make([]taskResult, len(jr.tasks))
	for index, taskExecution := range tasks {
		pathfile := constants.LOG_PATH_RUNNER_TASK(jr.schedulerName, jr.executeDatetime, taskExecution.GetName())
		jr.logger = logger.NewLoggerWithFile(pathfile)

		taskResult := taskResult{
			status:    constants.JOB_STATUS_RUNNING,
			startDate: time.Now(),
		}
		jr.logger.Info(fmt.Sprintf("scheduler %s with starting task %s", jr.schedulerName, taskExecution.GetName()), map[string]interface{}{"job_id": jr.id, "task_name": taskExecution.GetName(), "scheduler_name": jr.schedulerName, "task_start": taskResult.startDate.Format(constants.TIME_FORMAT_RFC339)})

		jr.currentTaskIndex = index

		/* save in db */
		saveJobTask(jr, taskExecution, taskResult)

		value, err := taskExecution.Call(jr.ctx)
		ti := time.Now()
		taskResult.endDatetime = &ti
		if err != nil {
			jr.exceptionOnTaskName = taskExecution.GetName()
			jr.exception = newRunnerException(err, false)
			taskResult.status = constants.JOB_STATUS_FAILED
			jr.taskResults[index] = taskResult
			jr.logger.Error(err, map[string]interface{}{"job_id": jr.id, "task_name": taskExecution.GetName(), "scheduler_name": jr.schedulerName, "task_start": taskResult.startDate.Format(constants.TIME_FORMAT_RFC339), "task_end": taskResult.endDatetime.Format(constants.TIME_FORMAT_RFC339), "task_status": taskResult.status})
			saveJobTask(jr, taskExecution, taskResult)
			return
		}
		taskResult.status = constants.JOB_STATUS_SUCCESS
		saveJobTask(jr, taskExecution, taskResult)

		jr.taskResults[index] = taskResult
		jr.taskValue.Store(taskExecution.GetName(), value)

		jr.logger.Info(fmt.Sprintf("scheduler %s with ending task %s", jr.schedulerName, taskExecution.GetName()), map[string]interface{}{"job_id": jr.id, "task_name": taskExecution.GetName(), "scheduler_name": jr.schedulerName, "task_start": taskResult.startDate.Format(constants.TIME_FORMAT_RFC339), "task_end": taskResult.endDatetime.Format(constants.TIME_FORMAT_RFC339), "task_status": taskResult.status})

		switch taskExecution.GetType() {
		case constants.TASK_TYPE_BRANCH_TASK:
			val := value.(task.TaskBranchPipeLine)
			tasks := val.GetTasks()
			jr.run(tasks)
			return
		}
	}

	jr.setStatus(constants.JOB_STATUS_SUCCESS)
}

func (jr *jobRunner) clear() {
	jr = nil
}

func (jr *jobRunner) setStartProcess() {
	jr.setStatus(constants.JOB_STATUS_RUNNING)
	jr.logjob.UpdatedAt = time.Now()
}

func (jr *jobRunner) setEndProcess() {
	ti := time.Now()
	jr.endDatetime = &ti
	jr.logjob.EndDatetime = &ti
	jr.logjob.UpdatedAt = ti
}

func (jr *jobRunner) setStatus(status constants.JobStatus) {
	jr.status = status
	jr.logjob.Status = jr.status
}
