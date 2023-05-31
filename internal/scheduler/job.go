package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/logger"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/go-co-op/gocron"
)

type JobInstance struct {
	scheduler     *SchedulerInstance
	Job           *gocron.Job
	tasks         []task.Execution
	arguments     map[string]interface{} // kargs of any process
	status        string
	totalTask     int
	logger        *logger.Log
	triggerConfig *sync.Map
}

func NewJob(arguments map[string]interface{}) *JobInstance {
	ji := &JobInstance{
		tasks:     make([]task.Execution, 0),
		arguments: make(map[string]interface{}),
		totalTask: 0,
	}
	if arguments != nil {
		ji.arguments = arguments
	}

	return ji
}

func (j *JobInstance) AddTask(tasks ...task.Execution) {
	j.tasks = append(j.tasks, tasks...)
	j.totalTask = len(j.tasks)
}

func (j JobInstance) GetTask(taskIndex int) task.Execution {
	return j.tasks[taskIndex]
}

func (j JobInstance) GetTotalTask() int {
	return j.totalTask
}

func (j *JobInstance) SetScheduler(scheudler *SchedulerInstance) {
	j.scheduler = scheudler
	if scheudler.config.JobMode == constants.JOB_MODE_SIGNLETON && scheudler.cronExpression != "" {
		j.Job.SingletonMode()
	}
	j.logger = logger.NewLoggerWithFile(constants.LOG_PATH_RESULT_JOB(scheudler.name))
}

func (j *JobInstance) trigger(overrideJobId string, triggerConfig *sync.Map, executeDatetime *time.Time) (jobId string, fn func()) {
	ctx := context.Background()
	runner := newJobRunner(ctx, j, triggerConfig, executeDatetime)
	runner.triggerType = constants.TRIGGER_TYPE_SCHEDULE
	if overrideJobId != "" {
		runner.id = overrideJobId
	}
	if executeDatetime != nil {
		runner.triggerType = constants.TRIGGER_TYPE_EXTERNAL
	}

	ctx = context.WithValue(ctx, constants.JOB_RUNNER_INSTANCE_KEY, runner.getRunnerInterface())
	runner.ctx = ctx

	runner.logjob = &models.Job{
		SchedulerName: j.scheduler.name,
		JobId:         runner.id,
		Status:        runner.status,
		StartDateTime: &runner.executeDatetime,
		EndDatetime:   nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := j.scheduler.GetAdapter().GetRepository().UpsertJob(ctx, runner.logjob); err != nil {
		fmt.Println("fail to upsert job with status WAITING:", err.Error())
	}
	return runner.id, func() {
		j.process(runner)
	}
}

func (j *JobInstance) process(runner *jobRunner) {
	if runner.triggerType == constants.TRIGGER_TYPE_SCHEDULE {
		trigger := &models.Trigger{
			SchedulerName:   j.scheduler.name,
			ExecuteDatetime: time.Now(),
			JobId:           runner.id,
			Config:          nil,
			TriggerType:     constants.TRIGGER_TYPE_SCHEDULE,
			IsTrigger:       true,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		j.scheduler.dbAdapter.GetRepository().UpsertTrigger(context.Background(), trigger)
	}

	runner.setStartProcess()
	if err := j.scheduler.GetAdapter().GetRepository().UpsertJob(runner.ctx, runner.logjob); err != nil {
		fmt.Println("fail to upsert job with status Before processing:", err.Error())
	}
	defer func() {
		runner.setEndProcess()
		if err := j.scheduler.GetAdapter().GetRepository().UpsertJob(runner.ctx, runner.logjob); err != nil {
			fmt.Println("fail to upsert job with status After processing:", err.Error())
		}
	}()

	defer runner.clear()
	runner.run(runner.tasks)
	if runner.exception != nil {
		if runner.logtaskrunning != nil {
			runner.logtaskrunning.Status = constants.JOB_STATUS_FAILED
			runner.logtaskrunning.UpdatedAt = time.Now()
			runner.logtaskrunning.TaskException = runner.exception.Error()
			runner.logtaskrunning.StackTrace = runner.exception.StackTrace()
			j.scheduler.GetAdapter().GetRepository().UpsertJobTask(runner.ctx, runner.logtaskrunning)
		}
		/* case error */
		if j.scheduler.config.OnError != nil {
			j.scheduler.config.OnError(runner.ctx)
		}

		// j.logger.Error(fmt.Errorf("error: scheduler on %s.%s with message %s", j.scheduler.name, runner.GetTask().GetName(), runner.exception.Error()), map[string]interface{}{
		// 	"job_id":             runner.id,
		// 	"scheduler_name":     j.scheduler.name,
		// 	"execution_datetime": runner.executeDatetime.Format(constants.TIME_FORMAT_RFC339),
		// 	"end_datetime":       time.Now().Format(constants.TIME_FORMAT_RFC339),
		// 	"status":             runner.GetStatus(),
		// 	"arguments":          constants.PARSE_SYNC_MAP_TO_MAP(runner.arguments),
		// 	"parameter":          constants.PARSE_SYNC_MAP_TO_MAP(runner.parameter),
		// })
		return
	}

	/* case success */
	if j.scheduler.config.OnSuccess != nil {
		j.scheduler.config.OnSuccess(runner.ctx)
	}
	runner.GetArguments()
	// j.logger.Info("dag result success", map[string]interface{}{
	// 	"job_id":             runner.id,
	// 	"scheduler_name":     j.scheduler.name,
	// 	"execution_datetime": runner.executeDatetime.Format(constants.TIME_FORMAT_RFC339),
	// 	"end_datetime":       runner.taskResults[len(runner.taskResults)-1].endDatetime.Format(constants.TIME_FORMAT_RFC339),
	// 	"status":             runner.GetStatus(),
	// 	"arguments":          constants.PARSE_SYNC_MAP_TO_MAP(runner.arguments),
	// 	"parameter":          constants.PARSE_SYNC_MAP_TO_MAP(runner.parameter),
	// })
}
