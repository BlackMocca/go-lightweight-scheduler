package scheduler

import (
	"context"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/go-co-op/gocron"
)

type JobInstance struct {
	Job             *gocron.Job
	tasks           []task.Execution
	arguments       map[string]interface{}
	status          string
	totalTask       int
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

// Override On Before
func (j *JobInstance) onBefore() {
	// ctx := j.Job.Context()
	runner := newJobRunner(context.Background(), j)

	runner.run()
	if runner.exception != nil {
		/* case error */
		j.schedulerConfig.OnError(runner.ctx)
		return
	}

	/* case success */
	j.schedulerConfig.OnSuccess(runner.ctx)
}
