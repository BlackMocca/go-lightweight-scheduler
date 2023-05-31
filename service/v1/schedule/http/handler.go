package http

import (
	"net/http"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/dag"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
)

type scheduleHandler struct {
	repository schedule.Repository
}

func NewScheduleHandler(repository schedule.Repository) schedule.HttpHandler {
	return &scheduleHandler{repository: repository}
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
	var ctx = c.Request().Context()
	var params = c.Get("params").(map[string]interface{})
	var name = cast.ToString(params["name"])
	var executeDatetime, _ = time.Parse(time.RFC3339, cast.ToString(params["execute_datetime"]))
	var config = params["config"]
	var schedule = sh.getOneSchedule(name)

	trigger := &models.Trigger{
		SchedulerName:   name,
		ExecuteDatetime: executeDatetime,
		IsActive:        true,
		IsTrigger:       false,
		TriggerType:     constants.TRIGGER_TYPE_EXTERNAL,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if config != nil {
		trigger.Config = config.(map[string]interface{})
	}

	jobId := schedule.Run(trigger)
	if err := sh.repository.UpsertTrigger(ctx, trigger); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"job_id": jobId,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) GetOneJobById(c echo.Context) error {
	var ctx = c.Request().Context()
	var jobId = c.Param("job_id")

	job, err := sh.repository.GetOneJob(ctx, jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if job == nil {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	trigger, err := sh.repository.GetOneTriggerByJobId(ctx, jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if trigger != nil {
		job.Trigger = trigger
	}

	jobtasks, err := sh.repository.GetOneJobTaskByJobId(ctx, jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(jobtasks) > 0 {
		job.JobRunningTasks = jobtasks
	}

	resp := map[string]interface{}{
		"job": job,
	}
	return c.JSON(http.StatusOK, resp)
}
