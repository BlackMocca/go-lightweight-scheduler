package dag

import (
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
)

func StartAllDag(stop chan bool) {
	schedulers := []*scheduler.SchedulerInstance{}
	if constants.ENV_ENABLED_DAG_EXAMPLE {
		schedulers = append(schedulers,
			startDagExampleGolang(),
			startDagExampleTaskBranch(),
			startDagExampleTaskBash(),
		)
	}

	for _, scheduler := range schedulers {
		scheduler.Start()
		fmt.Println("start scheduler:", scheduler.GetCronjobExpression(), scheduler.GetName())
		defer scheduler.Stop()
	}
	<-stop
}
