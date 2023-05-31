package route

import (
	"net/http"

	"github.com/Blackmocca/go-lightweight-scheduler/middleware"
	"github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule"
	_schedule_validator "github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule/validator"
	"github.com/labstack/echo/v4"
)

type Route struct {
	e     *echo.Echo
	middl middleware.RestAPIMiddleware
	auth  *echo.Group
}

func NewRoute(e *echo.Echo, middl middleware.RestAPIMiddleware) *Route {
	auth := e.Group("")
	auth.Use(middl.Authorization)

	return &Route{e: e, middl: middl, auth: auth}
}

func (r Route) RegisterHealthcheck() {
	r.e.GET("/healthcheck", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})
}

func (r Route) RegisterSchedule(handler schedule.HttpHandler, validation _schedule_validator.Validation) {
	r.auth.GET("/v1/schedulers", handler.GetListSchedule)
	r.auth.GET("/v1/scheduler/:name", handler.GetOneSchedule)
	r.auth.GET("/v1/job/:job_id", handler.GetOneJobById)
	r.auth.POST("/v1/scheduler/triggers", handler.Trigger, validation.ValidateTrigger)
}
