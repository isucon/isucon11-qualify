package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

var (
	catalogs map[string]*IsuCatalog
)

type IsuCatalog struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LimitWeight int64  `json:"limit_weight"`
	Weight      int64  `json:"weight"`
	Size        string `json:"size"`
	Maker       string `json:"maker"`
	Features    string `json:"features"`
}

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

func init() {
	catalogs = map[string]*IsuCatalog{
		"550e8400-e29b-41d4-a716-446655440000": {
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "isu0",
			LimitWeight: 150,
			Weight:      30,
			Size:        "W65.5×D66×H114.5~128.5cm",
			Maker:       "isu maker",
			Features:    "headrest,armrest",
		},
		"562dc0df-2d4f-4e38-98c0-9333f4ff3e38": {
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "isu1",
			LimitWeight: 136,
			Weight:      15,
			Size:        "W47×D43×H91cm～97cm",
			Maker:       "isu maker 2",
			Features:    "",
		},
	}
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
	catalogID := c.QueryParam("catalog_id")
	if catalogID == "" {
		// 全件取得
		catalogsArray := []*IsuCatalog{}
		for _, catalog := range catalogs {
			catalogsArray = append(catalogsArray, catalog)
		}
		return c.JSON(http.StatusOK, catalogsArray)
	}
	return c.JSON(http.StatusOK, catalogs[catalogID])
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
