package constants

import (
	"fmt"
	"time"
)

var (
	LOG_PATH_SCHEDULER = "./logs/scheduler.log"

	// dag/executedate/time/taskname.log
	LOG_PATH_RUNNER_TASK = func(schedulerName string, executeDatetime time.Time, taskName string) string {
		executeDate := executeDatetime.Format(TIME_FORMAT_DATE)
		ti := executeDatetime.Format(TIME_FORMAT_TIME)
		return fmt.Sprintf("./logs/%s/%s/%s/%s.log", schedulerName, executeDate, ti, taskName)
	}
)
