package example

import (
	"context"
	"errors"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
)

func run() {
	config := scheduler.NewDefaultSchedulerConfig()
	config.OnError = func(ctx context.Context) error {
		fmt.Println("on error")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(*scheduler.JobRunner)
		fmt.Println(jobRunner.GetTask().GetName())
		return nil
	}
	config.OnSuccess = func(ctx context.Context) error {
		fmt.Println("on success")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(*scheduler.JobRunner)

		parameter := jobRunner.GetParameter()
		fmt.Println(parameter)
		return nil
	}

	schedulerInstance := scheduler.NewScheduler("*/1 * * * *", "tmp", config)
	job := scheduler.NewJob(map[string]interface{}{"hi": "world"})
	job.AddTask(task.NewTask("testfoo1", executor.NewGolangExecuter(func(ctx context.Context) error {

		fmt.Println("this is foo1")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(*scheduler.JobRunner)

		parameter := jobRunner.GetParameter()
		parameter.Store("test", "bye")
		return errors.New("error test foo1")
	})))
	job.AddTask(task.NewTask("testfoo2", executor.NewGolangExecuter(func(ctx context.Context) error {
		fmt.Println("this is foo2")

		return nil
	})))
	schedulerInstance.RegisterJob(job)
	defer schedulerInstance.Stop()
}
