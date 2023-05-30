package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Blackmocca/go-lightweight-scheduler/dag"
	_ "github.com/Blackmocca/go-lightweight-scheduler/dag"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/middleware"
	"github.com/Blackmocca/go-lightweight-scheduler/route"
	_schedule_handler "github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule/http"
	_schedule_validator "github.com/Blackmocca/go-lightweight-scheduler/service/v1/schedule/validator"
	"github.com/labstack/echo/v4"
	echoMiddL "github.com/labstack/echo/v4/middleware"
)

func getWebInstance() (*echo.Echo, middleware.RestAPIMiddleware, *route.Route) {
	middL := middleware.NewRestAPIMiddleware()

	e := echo.New()
	e.Use(echoMiddL.Logger())
	e.Use(echoMiddL.Recover())
	e.Use(echoMiddL.AddTrailingSlash())
	e.Use(middL.InitContext)
	e.Use(middL.InputForm)

	router := route.NewRoute(e, middL)
	router.RegisterHealthcheck()

	schedulHandler := _schedule_handler.NewScheduleHandler()
	router.RegisterSchedule(schedulHandler, _schedule_validator.NewValidation())

	return e, middL, router
}

func main() {
	ctx := context.Background()

	// connection, err := connection.GetDatabaseConnection(ctx, constants.AdapterDatabaseConnectionType(constants.ENV_DATABASE_ADAPTER), constants.ENV_DATABASE_URL)
	// if err != nil {
	// 	panic(err)
	// }
	// defer connection.Close(ctx)

	e, _, _ := getWebInstance()

	go func() {
		port := fmt.Sprintf(":%s", constants.ENV_APP_PORT)
		e.Logger.Fatal(e.Start(port))
	}()

	stop := make(chan bool)
	go dag.StartAllDag(stop)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

	// graceful shutdown
	<-signalCh
	fmt.Println("cancel application")

	e.Shutdown(ctx)
	fmt.Println("shutdown web service")

	stop <- true
	fmt.Println("shutdown scheduler")
}
