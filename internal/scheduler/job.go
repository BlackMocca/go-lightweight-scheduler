package scheduler

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

type JobInstance struct {
	Job             *gocron.Job
	tasks           []task.Execution
	arguments       map[string]interface{}
	status          string
	totalTask       int
	schedulerName   string
	schedulerConfig SchedulerConfig // will be get config after register job
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

func (j *JobInstance) SetSchedulerConfig(config SchedulerConfig) {
	j.schedulerConfig = config
	if config.JobMode == constants.JOB_MODE_SIGNLETON {
		j.Job.SingletonMode()
	}
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
		j.schedulerConfig.OnError(runner.ctx)
		logrus.Error(fmt.Errorf("error: scheduler on %s.%s with message %s", j.schedulerName, runner.GetTask().GetName(), runner.exception.Error()))
		return
	}

	/* case success */
	j.schedulerConfig.OnSuccess(runner.ctx)
}
