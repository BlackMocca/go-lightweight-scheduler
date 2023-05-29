package dag

import "github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"

func StartAllDag(stop chan bool) {
	schedulers := []*scheduler.SchedulerInstance{
		startDagExampleGolang(),
		startDagTaskBranch(),
	}

	for _, scheduler := range schedulers {
		scheduler.Start()
		defer scheduler.Stop()
	}
	<-stop
}
