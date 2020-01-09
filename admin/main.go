package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	//fs := http.FileServer(http.Dir("/Users/baahk01/workspace/ugcuploader-test-kubernettes/admin/web"))
	//e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", fs)))
	e.Static("/static","web")
	e.File("/", "web/index.html")

	// Start server
	e.Logger.Debug(e.Start(":1323"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
