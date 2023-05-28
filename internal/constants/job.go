package constants

type JobMode int8

const (
	/*
		This mode will be waiting current process was complete and another
		process will be in queue
		queue must have only 100 queue
	*/
	JOB_MODE_SIGNLETON  JobMode = iota
	JOB_MODE_CONCURRENT JobMode = iota
)

type JobStatus string

const (
	JOB_STATUS_RUNNING JobStatus = "RUNNING"
	JOB_STATUS_SUCCESS JobStatus = "SUCCESS"
	JOB_STATUS_ERROR   JobStatus = "ERROR"
)

type JobContextKey string

const (
	JOB_RUNNER_INSTANCE_KEY JobContextKey = "instance" // using load current status of job runner
)
