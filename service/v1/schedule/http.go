package schedule

import "github.com/labstack/echo/v4"

type HttpHandler interface {
	GetListSchedule(echo.Context) error
	GetOneSchedule(echo.Context) error
	Trigger(echo.Context) error
	GetOneJobById(echo.Context) error
	GetListJob(echo.Context) error
	GetListJobTask(c echo.Context) error
	GetListJobFuture(c echo.Context) error
	UnActivatedTrigger(c echo.Context) error
	DeleteJobFuture(c echo.Context) error
}
