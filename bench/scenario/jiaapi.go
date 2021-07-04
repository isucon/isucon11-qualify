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
	streamsForPosterMutex sync.Mutex
	isuIsActivated        = map[string]JiaAPI2PosterData{}
	streamsForPoster      = map[string]*model.StreamsForPoster{}

	jiaAPIContext context.Context
	jiaAPIStep    *isucandar.BenchmarkStep
)

type IsuConditionPosterRequest struct {
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	IsuUUID    string `json:"isu_uuid"`
}

//ISU協会スレッドとposterの通信
type JiaAPI2PosterData struct {
	activated   bool
	chancelFunc context.CancelFunc
}

//シナリオスレッドからの呼び出し
func RegisterToJiaAPI(jiaIsuUUID string, streams *model.StreamsForPoster) {
	streamsForPosterMutex.Lock()
	defer streamsForPosterMutex.Unlock()
	streamsForPoster[jiaIsuUUID] = streams
}

func JiaAPIThread(ctx context.Context, step *isucandar.BenchmarkStep) {

	jiaAPIContext = ctx
	jiaAPIStep = step

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
	state := &IsuConditionPosterRequest{}
	err := c.Bind(state)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if !(0 <= state.TargetPort && state.TargetPort < 0x1000) {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	targetURL := fmt.Sprintf(
		"http://%s:%d/api/isu/%s/condition",
		state.TargetIP, state.TargetPort, state.IsuUUID,
	)

	//posterスレッドの起動
	posterContext, chancelFunc := context.WithCancel(jiaAPIContext)
	err = func() error {
		streamsForPosterMutex.Lock()
		defer streamsForPosterMutex.Unlock()
		scenarioChan, ok := streamsForPoster[state.IsuUUID]
		if !ok {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		v, ok := isuIsActivated[state.IsuUUID]
		if ok && v.activated {
			return echo.NewHTTPError(http.StatusForbidden)
		}
		isuIsActivated[state.IsuUUID] = JiaAPI2PosterData{
			activated:   true,
			chancelFunc: chancelFunc,
		}

		go KeepPosting(posterContext, jiaAPIStep, targetURL, scenarioChan)
		return nil
	}()
	if err != nil {
		return err
	}

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

	streamsForPosterMutex.Lock()
	defer streamsForPosterMutex.Unlock()
	v, ok := isuIsActivated[state.IsuUUID]
	if !(ok && v.activated) {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	v.chancelFunc()
	v.activated = false

	return c.NoContent(http.StatusNoContent)
}
