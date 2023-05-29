package dag

import (
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
)

var (
	SCHEDULERS = make([]*scheduler.SchedulerInstance, 0)
)

func StartAllDag(stop chan bool) {
	if constants.ENV_ENABLED_DAG_EXAMPLE {
		SCHEDULERS = append(SCHEDULERS,
			startDagExampleGolang(),
			startDagExampleTaskBranch(),
			startDagExampleTaskBash(),
		)
	}

	for _, scheduler := range SCHEDULERS {
		scheduler.Start()
		defer scheduler.Stop()
	}
	<-stop
}
