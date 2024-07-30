package http

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/Blackmocca/go-lightweight-scheduler/dag"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/models"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	"github.com/gofrs/uuid"
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
	if schedule == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("scheduler name '%s' not found", name))
	}

	uid, _ := uuid.NewV4()
	trigger := &models.Trigger{
		SchedulerName:   name,
		JobId:           uid.String(),
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

	if err := sh.repository.UpsertTrigger(ctx, trigger); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	jobId := schedule.Run(trigger)

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
		job = &models.Job{
			JobId: jobId,
		}
	}

	trigger, err := sh.repository.GetOneTriggerByJobId(ctx, jobId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if trigger != nil {
		job.Trigger = trigger
		if job.Status == "" {
			job.SchedulerName = trigger.SchedulerName
			job.Status = constants.JOB_STATUS_WAITING
		}
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

func (sh scheduleHandler) getArgs(c echo.Context) *sync.Map {
	var args = new(sync.Map)
	var searchWord = c.QueryParam("search_word")
	var status = c.QueryParam("status")
	var startDate = c.QueryParam("start_date")
	var endDate = c.QueryParam("end_date")

	if searchWord != "" {
		if _, err := uuid.FromString(searchWord); err == nil {
			args.Store("job_id", searchWord)
		} else {
			args.Store("search_word", searchWord)
		}
	}

	if status != "" {
		args.Store("status", status)
	}

	if status != "" {
		args.Store("status", status)
	}

	if startDate != "" {
		args.Store("start_date", startDate)
	}

	if endDate != "" {
		args.Store("end_date", endDate)
	}

	return args
}

func (sh scheduleHandler) getTotalPage(totalRow int, perPage int) int {
	return int(math.Ceil(float64(totalRow) / float64(perPage)))
}

func (sh scheduleHandler) GetListJob(c echo.Context) error {
	var ctx = c.Request().Context()
	var args = sh.getArgs(c)
	var page = 1
	var perPage = 20
	if querypage := c.QueryParam("page"); querypage != "" {
		page = cast.ToInt(querypage)
	}
	if queryperPage := c.QueryParam("per_page"); queryperPage != "" {
		perPage = cast.ToInt(queryperPage)
	}

	jobs, totalRow, err := sh.repository.GetJobs(ctx, args, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(jobs) == 0 {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	resp := map[string]interface{}{
		"jobs":       jobs,
		"page":       page,
		"per_page":   perPage,
		"total_page": sh.getTotalPage(totalRow, perPage),
		"total_row":  totalRow,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) GetListJobTask(c echo.Context) error {
	var ctx = c.Request().Context()
	var args = sh.getArgs(c)
	var page = 1
	var perPage = 20
	if querypage := c.QueryParam("page"); querypage != "" {
		page = cast.ToInt(querypage)
	}
	if queryperPage := c.QueryParam("per_page"); queryperPage != "" {
		perPage = cast.ToInt(queryperPage)
	}

	tasks, totalRow, err := sh.repository.GetJobTasks(ctx, args, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(tasks) == 0 {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	resp := map[string]interface{}{
		"tasks":      tasks,
		"page":       page,
		"per_page":   perPage,
		"total_page": sh.getTotalPage(totalRow, perPage),
		"total_row":  totalRow,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) GetListJobFuture(c echo.Context) error {
	var ctx = c.Request().Context()
	var args = sh.getArgs(c)
	var page = 1
	var perPage = 20
	if querypage := c.QueryParam("page"); querypage != "" {
		page = cast.ToInt(querypage)
	}
	if queryperPage := c.QueryParam("per_page"); queryperPage != "" {
		perPage = cast.ToInt(queryperPage)
	}

	triggers, totalRow, err := sh.repository.GetFutureJob(ctx, args, page, perPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(triggers) == 0 {
		return echo.NewHTTPError(http.StatusNoContent)
	}

	resp := map[string]interface{}{
		"triggers":   triggers,
		"page":       page,
		"per_page":   perPage,
		"total_page": sh.getTotalPage(totalRow, perPage),
		"total_row":  totalRow,
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) UnActivatedTrigger(c echo.Context) error {
	var ctx = c.Request().Context()
	var params = c.Get("params").(map[string]interface{})
	var name = cast.ToString(params["name"])
	var configKey = cast.ToString(params["config_key"])
	var configValue = params["config_value"]

	if err := sh.repository.UnActivatedTrigger(ctx, name, configKey, configValue); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"message": "successful",
	}
	return c.JSON(http.StatusOK, resp)
}

func (sh scheduleHandler) DeleteJobFuture(c echo.Context) error {
	var ctx = c.Request().Context()
	var jobId, err = uuid.FromString(c.Param("job_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "jobid must be uuid")
	}

	if err := sh.repository.UnActivatedTriggerByJobId(ctx, &jobId); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"message": "successful",
	}
	return c.JSON(http.StatusOK, resp)
}
