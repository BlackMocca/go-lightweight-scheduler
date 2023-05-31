package schedule

import "github.com/labstack/echo/v4"

type HttpHandler interface {
	GetListSchedule(echo.Context) error
	GetOneSchedule(echo.Context) error
	Trigger(echo.Context) error
	GetOneJobById(echo.Context) error
}
