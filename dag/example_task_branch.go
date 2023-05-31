package dag

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
)

// dag pipeline
// print1 >> taskExampleConsiderTaskRun return "branch1" >> print 3 >> print 4
// print1 >> taskExampleConsiderTaskRun return "branch2" >> print 2
// process ธรรมดา
func taskPrint1(ctx context.Context) (interface{}, error) {
	fmt.Println("print1")
	return nil, nil
}
func taskPrint2(ctx context.Context) (interface{}, error) {
	fmt.Println("print2")
	return nil, nil
}
func taskPrint3(ctx context.Context) (interface{}, error) {
	fmt.Println("print3")
	return nil, nil
}
func taskPrint4(ctx context.Context) (interface{}, error) {
	fmt.Println("print4")
	return nil, nil
}

// return task branch name ที่ต้องการจะไป
func taskExampleConsiderTaskRun(ctx context.Context) (interface{}, error) {
	return "branch3", nil
}

// return task branch name ที่ต้องการจะไป
func taskConsiderStack(ctx context.Context) (interface{}, error) {
	return "branch3.1", nil
}

func startDagExampleTaskBranch() *scheduler.SchedulerInstance {
	config := scheduler.NewDefaultSchedulerConfig()
	schedulerInstance := scheduler.NewScheduler("*/1 * * * *", "example_task_branch", config)

	job := scheduler.NewJob(nil)
	job.AddTask(
		task.NewTask("print1", executor.NewGolangExecuter(taskPrint1)),
		task.NewTaskBranch("taskExampleConsiderTaskRun", executor.NewGolangExecuter(taskExampleConsiderTaskRun), task.NewTaskBranchPipeline(map[string][]task.Execution{
			"branch1": {
				task.NewTask("print3", executor.NewGolangExecuter(taskPrint2)),
				task.NewTask("print4", executor.NewGolangExecuter(taskPrint4)),
			},
			"branch2": {
				task.NewTask("print3", executor.NewGolangExecuter(taskPrint3)),
			},
			"branch3": {
				task.NewTaskBranch("taskBrach3", executor.NewGolangExecuter(taskConsiderStack), task.NewTaskBranchPipeline(map[string][]task.Execution{
					"branch3.1": {
						task.NewTask("print2", executor.NewGolangExecuter(taskPrint2)),
						task.NewTask("print3", executor.NewGolangExecuter(taskPrint3)),
						task.NewTask("print4", executor.NewGolangExecuter(taskPrint4)),
					},
					"branch3.2": {
						task.NewTask("print2", executor.NewGolangExecuter(taskPrint2)),
						task.NewTask("print2", executor.NewGolangExecuter(taskPrint2)),
						task.NewTask("print2", executor.NewGolangExecuter(taskPrint2)),
					},
				})),
			},
		})),
	)

	schedulerInstance.RegisterJob(job)
	return schedulerInstance
}
