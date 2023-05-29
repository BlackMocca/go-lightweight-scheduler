package scheduler

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/logger"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/go-co-op/gocron"
)

type JobInstance struct {
	Job             *gocron.Job
	tasks           []task.Execution
	arguments       map[string]interface{}
	status          string
	totalTask       int
	schedulerName   string
	schedulerConfig SchedulerConfig // will be get config after register job
	logger          *logger.Log
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

func (j *JobInstance) SetScheduler(schedulerName string, config SchedulerConfig) {
	j.schedulerName = schedulerName
	j.schedulerConfig = config
	if config.JobMode == constants.JOB_MODE_SIGNLETON {
		j.Job.SingletonMode()
	}
	j.logger = logger.NewLoggerWithFile(constants.LOG_PATH_RESULT_JOB(schedulerName))
}

func (j *JobInstance) onBefore() {
	// ctx := j.Job.Context()
	ctx := context.Background()
	runner := newJobRunner(ctx, j)
	ctx = context.WithValue(ctx, constants.JOB_RUNNER_INSTANCE_KEY, runner.getRunnerInterface())
	runner.ctx = ctx

	defer runner.clear()
	runner.run(runner.tasks)
	if runner.exception != nil {
		/* case error */
		if j.schedulerConfig.OnError != nil {
			j.schedulerConfig.OnError(runner.ctx)
		}
		j.logger.Error(fmt.Errorf("error: scheduler on %s.%s with message %s", j.schedulerName, runner.GetTask().GetName(), runner.exception.Error()), map[string]interface{}{
			"jobId":              runner.id,
			"scheduler_name":     j.schedulerName,
			"execution_datetime": runner.executeDatetime.Format(constants.TIME_FORMAT_RFC339),
			"end_datetime":       runner.taskResults[len(runner.taskResults)-1].endDatetime.Format(constants.TIME_FORMAT_RFC339),
			"status":             runner.GetStatus(),
			"arguments":          constants.PARSE_SYNC_MAP_TO_MAP(runner.arguments),
			"parameter":          constants.PARSE_SYNC_MAP_TO_MAP(runner.parameter),
		})
		return
	}

	/* case success */
	if j.schedulerConfig.OnSuccess != nil {
		j.schedulerConfig.OnSuccess(runner.ctx)
	}
	runner.GetArguments()
	j.logger.Info("dag result success", map[string]interface{}{
		"jobId":              runner.id,
		"scheduler_name":     j.schedulerName,
		"execution_datetime": runner.executeDatetime.Format(constants.TIME_FORMAT_RFC339),
		"end_datetime":       runner.taskResults[len(runner.taskResults)-1].endDatetime.Format(constants.TIME_FORMAT_RFC339),
		"status":             runner.GetStatus(),
		"arguments":          constants.PARSE_SYNC_MAP_TO_MAP(runner.arguments),
		"parameter":          constants.PARSE_SYNC_MAP_TO_MAP(runner.parameter),
	})
}
