package controller

import (
	"net/http"
	"net/url"

	"github.com/isucon/isucon11-qualify/jiaapi-mock/model"
	"github.com/labstack/echo/v4"
)

/// Const Values ///
var (
	validIsu = map[string]ActivateResponse{
		"0694e4d7-dfce-4aec-b7ca-887ac42cfb8f": {Character: characterList[0]},
		"3a8ae675-3702-45b5-b1eb-1e56e96738ea": {Character: characterList[1]},
		"3efff0fa-75bc-4e3c-8c9d-ebfa89ecd15e": {Character: characterList[2]},
		"f67fcb64-f91c-4e7b-a48d-ddf1164194d0": {Character: characterList[3]},
		"32d1c708-e6ef-49d0-8ca9-4fd51844dcc8": {Character: characterList[4]},
		"f012233f-c50e-4349-9473-95681becff1e": {Character: characterList[5]},
		"af64735c-667a-4d95-a75e-22d0c76083e0": {Character: characterList[6]},
		"cb68f47f-25ef-46ec-965b-d72d9328160f": {Character: characterList[7]},
		"57d600ef-15b4-43bc-ab79-6399fab5c497": {Character: characterList[8]},
		"aa0844e6-812d-41d2-908a-eeb82a50b627": {Character: characterList[9]},
	}
	characterList = []string{
		"いじっぱり",
		"うっかりや",
		"おくびょう",
		"おだやか",
		"おっとり",
		"おとなしい",
		"がんばりや",
		"きまぐれ",
		"さみしがり",
		"しんちょう",
	}
)

func init() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 32
}

/// Struct for Req/Resp ///

type ActivateResponse struct {
	Character string `json:"character"`
}

type ActivationRequest struct {
	TargetBaseURL string `json:"target_base_url" validate:"required"`
	IsuUUID       string `json:"isu_uuid" validate:"required"`
}

/// Controller ///

type ActivationController struct {
	isuConditionPosterManager *model.IsuConditionPosterManager
}

func NewActivationController() *ActivationController {
	return &ActivationController{model.NewIsuConditionPosterManager()}
}

func (c *ActivationController) PostActivate(ctx echo.Context) error {
	req := &ActivationRequest{}
	err := ctx.Bind(req)
	if err != nil {
		ctx.Logger().Errorf("failed to bind: %v", err)
		return ctx.String(http.StatusBadRequest, "Bad Request")
	}

	parsedURL, err := url.Parse(req.TargetBaseURL)
	if err != nil {
		ctx.Logger().Errorf("bad url: %v", err)
		return ctx.String(http.StatusBadRequest, "Bad URL")
	}

	isuState, ok := validIsu[req.IsuUUID]
	if !ok {
		ctx.Logger().Errorf("bad isu_uuid: %v", req.IsuUUID)
		return ctx.String(http.StatusNotFound, "Bad isu_uuid")
	}

	err = c.isuConditionPosterManager.StartPosting(parsedURL, req.IsuUUID)
	if err != nil {
		ctx.Logger().Errorf("failed to startPosting: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusAccepted, isuState)
}
