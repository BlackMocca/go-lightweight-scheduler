package main

import (
	"context"
	"fmt"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/connection"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/Blackmocca/go-lightweight-scheduler/internal/scheduler"
	"github.com/Blackmocca/go-lightweight-scheduler/middleware"
	"github.com/Blackmocca/go-lightweight-scheduler/route"
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

	return e, middL, router
}

func main() {
	ctx := context.Background()

	connection, err := connection.GetDatabaseConnection(ctx, constants.AdapterDatabaseConnectionType(constants.ENV_DATABASE_ADAPTER), constants.ENV_DATABASE_URL)
	if err != nil {
		panic(err)
	}
	defer connection.Close(ctx)

	e, _, _ := getWebInstance()

	fmt.Println(connection)

	schedulerInstance := scheduler.NewScheduler("*/1 * * * *", "tmp", scheduler.NewDefaultSchedulerConfig())
	schedulerInstance.RegisterJob(scheduler.NewJob("tmp", nil))
	schedulerInstance.Start()
	defer schedulerInstance.Stop()
	// schedulerInstance.Scheduler.Every(1).Seconds().Do(func() {
	// 	fmt.Println("test")
	// })

	port := fmt.Sprintf(":%s", constants.ENV_APP_PORT)
	e.Logger.Fatal(e.Start(port))
}
