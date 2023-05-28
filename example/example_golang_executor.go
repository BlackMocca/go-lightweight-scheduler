package example

import (
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
)

func get() {

}

func run() {
	schedulerInstance := scheduler.NewScheduler("*/1 * * * *", "example_golang_excutor", scheduler.NewDefaultSchedulerConfig())
	schedulerInstance.RegisterJob(scheduler.NewJob("tmp", nil))
	schedulerInstance.Start()
	defer schedulerInstance.Stop()
	schedulerInstance.Scheduler.Every(1).Seconds().Do(func() {
		fmt.Println("test")
	})
}
