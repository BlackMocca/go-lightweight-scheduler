package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
	"github.com/go-co-op/gocron"
)

type JobInstance struct {
	Job             *gocron.Job
	Id              string
	Name            string
	tasks           []task.Execution
	arguments       *sync.Map
	status          string
	totalTask       int
	schedulerConfig SchedulerConfig // will be get config after register job
}

var index = 0

func (j *JobInstance) onBefore() {
	ctx := j.Job.Context()
	if index == 0 {
		j.status = "test"
	}
	index++
	fmt.Println("test scope status jobInstance", j.status)
	// ctx.Value("runner")

	if j.schedulerConfig.onBefore != nil {
		if err := j.schedulerConfig.onBefore(ctx); err != nil {
			panic(err)
		}
	}
	for _, task := range j.tasks {
		if err := task.Call(ctx); err != nil {
			panic(err)
		}
	}
}

func (j *JobInstance) onAfter() {
	ctx := j.Job.Context()
	if j.schedulerConfig.onAfter != nil {
		if err := j.schedulerConfig.onAfter(ctx); err != nil {
			panic(err)
		}
	}
}

func NewJob(jobId string, arguments *sync.Map) *JobInstance {
	ji := &JobInstance{
		Id:        jobId,
		tasks:     make([]task.Execution, 0),
		arguments: new(sync.Map),
		// currentTaskIndex: 0,
		totalTask: 0,
	}
	if arguments != nil {
		ji.arguments = arguments
	}

	return ji
}

func (j *JobInstance) AddTask(...task.Execution) {
	j.tasks = append(j.tasks)
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

func (j *JobInstance) setEventListeners(onBefore, onAfter, onSuccess, onError func(ctx context.Context) error) {

}
