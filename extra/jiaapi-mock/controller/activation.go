package controller

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/isucon/isucon11-qualify/jiaapi-mock/model"
	"github.com/labstack/echo/v4"
)

/// Const Values ///
var (
	validIsu        map[string]ActivateResponse
	characterList   []string
	privateIPBlocks []*net.IPNet
)

func init() {
	characterList = []string{"Normal", "clean freak"}
	validIsu = map[string]ActivateResponse{
		"0694e4d7-dfce-4aec-b7ca-887ac42cfb8f": {Character: characterList[0]},
		"3a8ae675-3702-45b5-b1eb-1e56e96738ea": {Character: characterList[1]},
		"3efff0fa-75bc-4e3c-8c9d-ebfa89ecd15e": {Character: characterList[0]},
		"f67fcb64-f91c-4e7b-a48d-ddf1164194d0": {Character: characterList[1]},
		"32d1c708-e6ef-49d0-8ca9-4fd51844dcc8": {Character: characterList[0]},
		"f012233f-c50e-4349-9473-95681becff1e": {Character: characterList[1]},
		"af64735c-667a-4d95-a75e-22d0c76083e0": {Character: characterList[0]},
		"cb68f47f-25ef-46ec-965b-d72d9328160f": {Character: characterList[1]},
		"57d600ef-15b4-43bc-ab79-6399fab5c497": {Character: characterList[0]},
		"aa0844e6-812d-41d2-908a-eeb82a50b627": {Character: characterList[1]},
	}

	//privateIPBlocks
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}

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
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	parsedURL, err := url.Parse(req.TargetBaseURL)
	if err != nil {
		ctx.Logger().Errorf("bad url: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	split := strings.Split(parsedURL.Host, ":")
	if len(split) != 2 {
		ctx.Logger().Errorf("bad url")
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	host := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		ctx.Logger().Errorf("bad url: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if !(0 <= port && port < 0x1000) {
		ctx.Logger().Errorf("bad port: %v", port)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	isuState, ok := validIsu[req.IsuUUID]
	if !ok {
		ctx.Logger().Errorf("bad isu_uuid: %v", req.IsuUUID)
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if !isPrivateIP(host) {
		ctx.Logger().Errorf("bad ip: %v", host)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	err = c.isuConditionPosterManager.StartPosting(req.TargetBaseURL, req.IsuUUID)
	if err != nil {
		ctx.Logger().Errorf("failed to startPosting: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusAccepted, isuState)
}

func isPrivateIP(ipstr string) bool {

	ipAddr, err := net.ResolveIPAddr("ip", ipstr)
	if err != nil || ipAddr == nil {
		return false
	}
	ip := ipAddr.IP

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
