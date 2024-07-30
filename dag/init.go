package dag

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/connection"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
)

var (
	SCHEDULERS = make([]*scheduler.SchedulerInstance, 0)
)

func call() {
	if constants.ENV_ENABLED_DAG_EXAMPLE {
		startDagExampleGolang()
		startDagExampleTaskBash()
		startDagExampleWorkWithoutCronjob()
		startDagExampleTaskBranch()
	}
	//startdagExampleNewbie()
}

func register(scheduler *scheduler.SchedulerInstance) {
	SCHEDULERS = append(SCHEDULERS, scheduler)
}

func StartAllDag(stop chan bool, adapterConnection connection.DatabaseAdapterConnection) {
	call()
	for _, scheduler := range SCHEDULERS {
		scheduler.SetAdapter(adapterConnection)
		if err := scheduler.Start(); err != nil {
			panic(fmt.Sprintf("failed to start scheduler %s: %s", scheduler.GetName(), err.Error()))
		}
		defer scheduler.Stop()

		triggers, err := adapterConnection.GetRepository().GetTriggerTimer(context.Background(), scheduler.GetName())
		if err != nil {
			panic(err)
		}
		if len(triggers) > 0 {
			for _, trigger := range triggers {
				scheduler.Run(trigger)
				fmt.Println("load scheduler timer with job", trigger.JobId)
			}
		}
	}

	<-stop
}
