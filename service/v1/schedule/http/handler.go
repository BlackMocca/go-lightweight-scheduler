package http

import (
	"net/http"

	"github.com/Blackmocca/go-lightweight-scheduler/dag"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	"github.com/labstack/echo/v4"
)

type scheduleHandler struct {
}

func NewScheduleHandler() schedule.HttpHandler {
	return &scheduleHandler{}
}

func (sh scheduleHandler) GetListSchedule(c echo.Context) error {
	resp := map[string]interface{}{
		"schedulers": dag.SCHEDULERS,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) GetOneSchedule(c echo.Context) error {
	var name = c.Param("name")
	var schedule *scheduler.SchedulerInstance

	for _, item := range dag.SCHEDULERS {
		if name == item.GetName() {
			schedule = item
			break
		}
	}

	if schedule == nil {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	resp := map[string]interface{}{
		"scheduler": schedule,
	}
	return c.JSON(http.StatusOK, resp)
}
