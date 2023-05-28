package scheduler

import "github.com/Blackmocca/go-lightweight-scheduler/internal/constants"

type JobRunner struct {
	JobInstance
	status           constants.JobStatus
	currentTaskIndex int
}

func newJobRunner(ji JobInstance) JobRunner {
	return JobRunner{
		JobInstance: ji,
		status:      constants.JOB_STATUS_RUNNING,
	}
}
