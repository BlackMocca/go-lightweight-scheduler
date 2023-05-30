package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/dag"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
)

type scheduleHandler struct {
}

func NewScheduleHandler() schedule.HttpHandler {
	return &scheduleHandler{}
}

func (sh scheduleHandler) getOneSchedule(name string) *scheduler.SchedulerInstance {
	var schedule *scheduler.SchedulerInstance

	for _, item := range dag.SCHEDULERS {
		if name == item.GetName() {
			schedule = item
			break
		}
	}

	return schedule
}

func (sh scheduleHandler) GetListSchedule(c echo.Context) error {
	resp := map[string]interface{}{
		"schedulers": dag.SCHEDULERS,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) GetOneSchedule(c echo.Context) error {
	var name = c.Param("name")
	var schedule = sh.getOneSchedule(name)

	if schedule == nil {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	resp := map[string]interface{}{
		"scheduler": schedule,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) Trigger(c echo.Context) error {
	var params = c.Get("params").(map[string]interface{})
	var name = cast.ToString(params["name"])
	var executeDatetime, _ = time.Parse(time.RFC3339, cast.ToString(params["execute_datetime"]))
	var config = params["config"]
	var schedule = sh.getOneSchedule(name)

	var triggerConfig *sync.Map
	if config != nil {
		triggerConfig = new(sync.Map)
		for k, v := range config.(map[string]interface{}) {
			triggerConfig.Store(k, v)
		}
	}

	jobId := schedule.Run(triggerConfig, executeDatetime)

	resp := map[string]interface{}{
		"scheduler": schedule,
		"jobId":     jobId,
	}
	return c.JSON(http.StatusOK, resp)
}
