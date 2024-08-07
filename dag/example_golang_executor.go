package dag

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
)

func startDagExampleGolang() {
	config := scheduler.NewDefaultSchedulerConfig()
	config.OnError = func(ctx context.Context) error {
		fmt.Println("on error")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(scheduler.JobRunner)
		fmt.Println("task", jobRunner.GetTask().GetName(), "with exception", jobRunner.GetException().Error())
		return nil
	}
	config.OnSuccess = func(ctx context.Context) error {
		fmt.Println("on success")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(scheduler.JobRunner)

		parameter := jobRunner.GetParameter()
		fmt.Println(parameter)
		return nil
	}
	schedulerInstance := scheduler.NewScheduler("*/5 * * * *", "example_golang_executor", "ทดสอบ task func by golang", config)

	job := scheduler.NewJob(map[string]interface{}{"hi": "world"})
	job.AddTask(task.NewTask("testfoo1", executor.NewGolangExecuter(func(ctx context.Context) (interface{}, error) {

		fmt.Println("this is foo1")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(scheduler.JobRunner)

		config := jobRunner.GetTriggerConfig()
		fmt.Println(config)

		parameter := jobRunner.GetParameter()
		parameter.Store("test", "bye")

		return nil, nil
	})))
	job.AddTask(task.NewTask("testfoo2", executor.NewGolangExecuter(func(ctx context.Context) (interface{}, error) {
		fmt.Println("this is foo2")
		val := ctx.Value(constants.JOB_RUNNER_INSTANCE_KEY)
		jobRunner := val.(scheduler.JobRunner)

		arguments := jobRunner.GetArguments()
		parameter := jobRunner.GetParameter()
		triggerConfig := jobRunner.GetTriggerConfig()
		fmt.Println("arguments", arguments)
		fmt.Println("parameter", parameter)
		fmt.Println("triggerconfig", triggerConfig)

		return nil, nil
	})))

	schedulerInstance.RegisterJob(job)
	register(schedulerInstance)
}
