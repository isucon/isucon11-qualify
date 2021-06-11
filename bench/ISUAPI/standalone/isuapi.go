package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

func init() {
}

func main() {
	// Echo instance
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.GET("/api/catalog", getCatalog)
	e.POST("/api/activate", postActivate)
	e.POST("/api/deactivate", postDeactivate)
	e.POST("/api/die", postDie)

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("ISUAPI_SERVER_PORT", "5000"))
	e.Logger.Fatal(e.Start(serverPort))
}

func getCatalog(c echo.Context) error {
	return fmt.Errorf("not implemented")
}

func postActivate(c echo.Context) error {
	return fmt.Errorf("not implemented")
}

func postDeactivate(c echo.Context) error {
	return fmt.Errorf("not implemented")
}

func postDie(c echo.Context) error {
	os.Exit(0)
	return nil
}
