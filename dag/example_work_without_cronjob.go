package dag

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
)

func startDagExampleWorkWithoutCronjob() {
	config := scheduler.NewDefaultSchedulerConfig()
	schedulerInstance := scheduler.NewScheduler("", "example_work_without_cronjob", "ทดสอบ bash_executor", config)

	job := scheduler.NewJob(nil)
	job.AddTask(
		task.NewTask("test1", executor.NewGolangExecuter(func(ctx context.Context) (interface{}, error) {
			val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
			jobRunner := val.(scheduler.JobRunner)

			config := jobRunner.GetTriggerConfig()
			fmt.Println(config)

			return nil, nil
		})),
	)

	schedulerInstance.RegisterJob(job)
	register(schedulerInstance)
}
