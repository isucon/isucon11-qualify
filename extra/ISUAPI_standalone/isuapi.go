package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

var (
	privateIPBlocks []*net.IPNet

	catalogs        map[string]*IsuCatalog
	validIsu        map[string]IsuState
	characterList   []string
	activatedIsu    = map[string]ActivatedIsuState{} //key=isuID+targetIP
	activatedIsuMtx = sync.Mutex{}
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

type IsuState struct {
	CatalogID string `json:"catalog_id"`
	Character string `json:"character"`
}
type ActivatedIsuState struct {
	cancelFunc context.CancelFunc
}

type isuConditionPoster struct {
	targetIP   string
	targetPort int
	isuID      string
}
type IsuCondition struct {
	IsDirty      bool `json:"is_dirty"`
	IsOverweight bool `json:"is_overweight"`
	IsBroken     bool `json:"is_broken"`
}
type IsuNotification struct {
	IsSitting bool         `json:"is_sitting"`
	Condition IsuCondition `json:"condition"`
	Message   string       `json:"message"`
	Timestamp string       `json:"timestamp"`
}

func getEnv(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}
func isPrivateIP(ipstr string) bool {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return false
	}

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
	characterList = []string{"Normal", "clean freak"}
	validIsu = map[string]IsuState{
		"0694e4d7-dfce-4aec-b7ca-887ac42cfb8f": {CatalogID: "550e8400-e29b-41d4-a716-446655440000", Character: characterList[0]},
		"3a8ae675-3702-45b5-b1eb-1e56e96738ea": {CatalogID: "550e8400-e29b-41d4-a716-446655440000", Character: characterList[1]},
		"3efff0fa-75bc-4e3c-8c9d-ebfa89ecd15e": {CatalogID: "550e8400-e29b-41d4-a716-446655440000", Character: characterList[0]},
		"f67fcb64-f91c-4e7b-a48d-ddf1164194d0": {CatalogID: "550e8400-e29b-41d4-a716-446655440000", Character: characterList[1]},
		"32d1c708-e6ef-49d0-8ca9-4fd51844dcc8": {CatalogID: "550e8400-e29b-41d4-a716-446655440000", Character: characterList[0]},
		"f012233f-c50e-4349-9473-95681becff1e": {CatalogID: "562dc0df-2d4f-4e38-98c0-9333f4ff3e38", Character: characterList[1]},
		"af64735c-667a-4d95-a75e-22d0c76083e0": {CatalogID: "562dc0df-2d4f-4e38-98c0-9333f4ff3e38", Character: characterList[0]},
		"cb68f47f-25ef-46ec-965b-d72d9328160f": {CatalogID: "562dc0df-2d4f-4e38-98c0-9333f4ff3e38", Character: characterList[1]},
		"57d600ef-15b4-43bc-ab79-6399fab5c497": {CatalogID: "562dc0df-2d4f-4e38-98c0-9333f4ff3e38", Character: characterList[0]},
		"aa0844e6-812d-41d2-908a-eeb82a50b627": {CatalogID: "562dc0df-2d4f-4e38-98c0-9333f4ff3e38", Character: characterList[1]},
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

func main() {
	// Echo instance
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.GET("/api/catalog/:catalog_id", getCatalog)
	e.POST("/api/activate", postActivate)
	e.POST("/api/deactivate", postDeactivate)
	e.POST("/api/die", postDie)

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("ISUAPI_SERVER_PORT", "5000"))
	e.Logger.Fatal(e.Start(serverPort))
}

func getCatalog(c echo.Context) error {
	catalogID := c.Param("catalog_id")
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
	isuID := c.FormValue("isu_id")
	if isuID == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	targetIP := c.FormValue("target_ip")
	if targetIP == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	targetPort, err := strconv.Atoi(c.FormValue("target_port"))
	if err != nil || !(0 <= targetPort && targetPort < 0x1000) {
		return c.NoContent(http.StatusBadRequest)
	}

	state := &isuConditionPoster{
		targetIP:   targetIP,
		targetPort: targetPort,
		isuID:      isuID,
	}
	if _, ok := validIsu[state.isuID]; !ok {
		return c.NoContent(http.StatusNotFound)
	}
	if !isPrivateIP(state.targetIP) {
		return c.NoContent(http.StatusForbidden)
	}
	key := state.isuID + state.targetIP + strconv.Itoa(state.targetPort)

	ctx, cancel := context.WithCancel(context.Background())
	conflict := func() bool {
		activatedIsuMtx.Lock()
		defer activatedIsuMtx.Unlock()
		if _, ok := activatedIsu[key]; ok {
			return true
		}
		activatedIsu[key] = ActivatedIsuState{cancelFunc: cancel}
		return false
	}()
	if !conflict {
		go state.keepPosting(ctx)
	}

	return c.JSON(http.StatusAccepted, validIsu[isuID])
}

func postDeactivate(c echo.Context) error {
	isuID := c.FormValue("isu_id")
	if isuID == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	if _, ok := validIsu[isuID]; !ok {
		return c.NoContent(http.StatusNotFound)
	}
	targetIP := c.FormValue("target_ip")
	if targetIP == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	targetPort, err := strconv.Atoi(c.FormValue("target_port"))
	if err != nil || !(0 <= targetPort && targetPort < 0x1000) {
		return c.NoContent(http.StatusBadRequest)
	}

	key := isuID + targetIP + strconv.Itoa(targetPort)
	func() {
		activatedIsuMtx.Lock()
		defer activatedIsuMtx.Unlock()
		activatedState, ok := activatedIsu[key]
		if ok {
			activatedState.cancelFunc()
			delete(activatedIsu, key)
		}
	}()

	return c.NoContent(http.StatusNoContent)
}

func postDie(c echo.Context) error {
	password := c.FormValue("password")
	if password == "U,YaCLe9tAnW8EdYphW)Wc/dN)5pPQ/3ue_af4rz" {
		os.Exit(0)
	}
	return echo.NewHTTPError(http.StatusNotFound, "Not Found")
}

func (state *isuConditionPoster) keepPosting(ctx context.Context) {
	targetURL := fmt.Sprintf(
		"http://%s:%d/api/isu/%s/condition",
		state.targetIP, state.targetPort, state.isuID,
	)
	randEngine := rand.New(rand.NewSource(0))

	timer := time.NewTicker(2 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		//乱数初期化（逆算できるように）
		nowTime := time.Now()
		randEngine.Seed(nowTime.UnixNano()/1000000000 + 961054102)

		notification, err := json.Marshal(IsuNotification{
			IsSitting: true,
			Condition: IsuCondition{
				IsDirty:      (randEngine.Intn(2) == 0),
				IsOverweight: (randEngine.Intn(2) == 0),
				IsBroken:     (randEngine.Intn(2) == 0),
			},
			Message:   "今日もいい天気",
			Timestamp: nowTime.Format("2006-01-02 15:04:05 -0700"),
		})
		if err != nil {
			log.Error(err)
			continue
		}

		func() {
			resp, err := http.Post(
				targetURL, "application/json",
				bytes.NewBuffer(notification),
			)
			if err != nil {
				log.Error(err)
				return // goto next loop
			}
			defer resp.Body.Close()

			if resp.StatusCode != 201 {
				log.Errorf("failed to `POST %s` with status=`%s`", targetURL, resp.Status)
				return // goto next loop
			}
		}()
	}
}
