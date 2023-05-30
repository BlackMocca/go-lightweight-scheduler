package dag

import (
	"github.com/Blackmocca/go-lightweight-scheduler/internal/executor"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/task"
)

func startDagExampleTaskBash() *scheduler.SchedulerInstance {
	config := scheduler.NewDefaultSchedulerConfig()
	schedulerInstance := scheduler.NewScheduler("*/5 * * * *", "example_bash_executor", config)

	job := scheduler.NewJob(nil)
	job.AddTask(
		task.NewTask("runbash", executor.NewBashExecutor(`ls -al`, true)),
		task.NewTask("runscript", executor.NewBashExecutor(`./script/test_bash.sh`, true)),
	)

	schedulerInstance.RegisterJob(job)
	return schedulerInstance
}
