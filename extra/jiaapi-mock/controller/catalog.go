package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

/// Const Values ///

var (
	catalogs map[string]*getCatalogResponse
)

func init() {
	catalogs = map[string]*getCatalogResponse{
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

/// Struct for Req/Resp ///

type getCatalogResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LimitWeight int64  `json:"limit_weight"`
	Weight      int64  `json:"weight"`
	Size        string `json:"size"`
	Maker       string `json:"maker"`
	Features    string `json:"features"`
}

/// Controller ///

type CatalogController struct{}

func (c CatalogController) GetCatalog(ctx echo.Context) error {
	catalogID := ctx.Param("catalog_id")
	catalog, ok := catalogs[catalogID]
	if !ok {
		ctx.Logger().Errorf("bad catalog_id: %v", catalogID)
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return ctx.JSON(http.StatusOK, catalog)
}
