package main

import (
	AuthRouter "cms/router/auth"
	//MainRouter "cms/router/mains"
	"cms/db"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var e *echo.Echo
var api *echo.Group

func main() {
	e = echo.New()
	db.InitDb()
	InitRout()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))

	e.Logger.Fatal(e.Start(":8123"))
}

func InitRout() {
	auth := e.Group("/api/v1/auth")
	AuthRouter.InitAuth(auth)

	//api = e.Group("/api/v1/user")
	//MainRouter.InitApi(api)
}
