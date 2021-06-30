package scenario

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	isuPosterMutex sync.Mutex
	isuIsActivated = map[string]IsuAPI2PosterData{}
	isuPosterData  = map[string]*model.IsuPosterChan{}
)

type IsuConditionPosterRequest struct {
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	IsuUUID    string `json:"isu_uuid"`
}

//ISU協会スレッドとposterの通信
type IsuAPI2PosterData struct {
	activated   bool
	chancelFunc context.CancelFunc
}

//シナリオスレッドからの呼び出し
func RegisterToIsuAPI(data *model.IsuPosterChan) {
	isuPosterMutex.Lock()
	defer isuPosterMutex.Unlock()
	isuPosterData[data.JIAIsuUUID] = data
}

func IsuAPIThread(ctx context.Context, step *isucandar.BenchmarkStep) {

	// Echo instance
	e := echo.New()
	//e.Debug = true
	//e.Logger.SetLevel(log.DEBUG)

	// Middleware
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.GET("/api/catalog/:catalog_id", getCatalog)
	e.POST("/api/activate", postActivate)
	e.POST("/api/deactivate", postDeactivate)

	// Start server
	go func() {
		serverPort := ":80"
		err := e.Start(serverPort)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("ISU協会サービスが異常終了しました: %v", err))
		}
	}()

	//コンテキストにより終了された場合は、echoサーバーも終了
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := e.Shutdown(ctx)
	if err != nil {
		//有効なエラー処理は出来ないのでエラーは握り潰し
		logger.AdminLogger.Printf("Failed to write prom file: %s", err)
	}
}

func getCatalog(c echo.Context) error {
	//TODO:
	return fmt.Errorf("NotImplemented")
}

func postActivate(c echo.Context) error {
	//state := &IsuConditionPosterRequest{}
	// err := c.Bind(state)
	// if err != nil {
	// 	c.Logger().Errorf("failed to bind: %v", err)
	// 	return echo.NewHTTPError(http.StatusBadRequest)
	// }
	// if !(0 <= state.TargetPort && state.TargetPort < 0x1000) {
	// 	c.Logger().Errorf("bad port: %v", state.TargetPort)
	// 	return echo.NewHTTPError(http.StatusBadRequest)
	// }

	// isuState, ok := validIsu[state.IsuUUID]
	// if !ok {
	// 	c.Logger().Errorf("bad isu_uuid: %v", state.IsuUUID)
	// 	return echo.NewHTTPError(http.StatusNotFound)
	// }
	// if !isPrivateIP(state.TargetIP) {
	// 	c.Logger().Errorf("bad ip: %v", state.TargetIP)
	// 	return echo.NewHTTPError(http.StatusForbidden)
	// }

	// err = state.startPosting()
	// if err != nil {
	// 	c.Logger().Errorf("failed to startPosting: %v", err)
	// 	return echo.NewHTTPError(http.StatusInternalServerError)
	// }

	//return c.JSON(http.StatusAccepted, isuState)
	//TODO:
	return fmt.Errorf("NotImplemented")
}

func postDeactivate(c echo.Context) error {
	state := &IsuConditionPosterRequest{}
	err := c.Bind(state)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if !(0 <= state.TargetPort && state.TargetPort < 0x1000) {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	isuPosterMutex.Lock()
	defer isuPosterMutex.Unlock()
	v, ok := isuIsActivated[state.IsuUUID]
	if !(ok && v.activated) {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	v.chancelFunc()
	v.activated = false

	return c.NoContent(http.StatusNoContent)
}
