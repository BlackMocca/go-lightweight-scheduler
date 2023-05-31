package constants

import (
	"time"
)

var (
	offset      = time.Second * time.Duration(7*60*60)
	TIME_CHANGE = func(ti time.Time, isChangeOnlyUTC bool) time.Time {
		ti = ti.Local()
		if isChangeOnlyUTC {
			ti = ti.Add(-time.Duration(offset))
		}
		return ti
	}
)

const (
	TIME_FORMAT_DATE   = "2006-01-02"
	TIME_FORMAT_TIME   = "15:04:05"
	TIME_FORMAT_RFC339 = time.RFC3339
)
