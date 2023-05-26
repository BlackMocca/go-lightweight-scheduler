package route

import (
	"net/http"

	"github.com/Blackmocca/go-lightweight-scheduler/middleware"
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
