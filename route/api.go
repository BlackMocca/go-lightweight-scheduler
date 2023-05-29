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
}

func NewRoute(e *echo.Echo, middl middleware.RestAPIMiddleware) *Route {
	return &Route{e: e, middl: middl}
}

func (r Route) RegisterHealthcheck() {
	r.e.GET("/healthcheck", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})
}

func (r Route) RegisterSchedule(handler schedule.HttpHandler, validation _schedule_validator.Validation) {
	r.e.GET("/v1/schedulers", handler.GetListSchedule)
	r.e.GET("/v1/scheduler/:name", handler.GetOneSchedule)
}
