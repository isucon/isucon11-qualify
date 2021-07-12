package scenario

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
	isuDetailInfomation   = map[string]*IsuDetailInfomation{}

	jiaAPIContext context.Context
	jiaAPIStep    *isucandar.BenchmarkStep
)

type IsuDetailInfomation struct {
	CatalogID string `json:"catalog_id"`
	Character string `json:"character"`
}

type IsuConditionPosterRequest struct {
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	IsuUUID    string `json:"isu_uuid"`
}

//ISU協会 Goroutineとposterの通信
type JiaAPI2PosterData struct {
	activated  bool
	cancelFunc context.CancelFunc
}

//シナリオ Goroutineからの呼び出し
func RegisterToJiaAPI(jiaIsuUUID string, detail *IsuDetailInfomation, streams *model.StreamsForPoster) {
	streamsForPosterMutex.Lock()
	defer streamsForPosterMutex.Unlock()
	isuDetailInfomation[jiaIsuUUID] = detail
	streamsForPoster[jiaIsuUUID] = streams
}

func (s *Scenario) JiaAPIService(ctx context.Context, step *isucandar.BenchmarkStep) {
	defer logger.AdminLogger.Println("--- JiaAPIService END")

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
	e.GET("/api/catalog/:catalog_id", func(c echo.Context) error { return s.getCatalog(c) })
	e.POST("/api/activate", func(c echo.Context) error { return s.postActivate(c) })
	e.POST("/api/deactivate", postDeactivate)

	// Start
	serverPort := s.jiaServiceURL[strings.LastIndexAny(s.jiaServiceURL, ":"):] //":80"
	s.loadWaitGroup.Add(1)
	go func() {
		defer logger.AdminLogger.Println("--- ISU協会サービス END")
		defer s.loadWaitGroup.Done()
		err := e.Start(serverPort)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("ISU協会サービスが異常終了しました: %v", err))
		}
	}()

	//コンテキストにより終了された場合は、echoサーバーも終了
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := e.Shutdown(ctx)
	if err != nil {
		//有効なエラー処理は出来ないのでエラーは握り潰し
		logger.AdminLogger.Printf("Failed to shutdown jia service: %s", err)
	}
}

func (s *Scenario) getCatalog(c echo.Context) error {
	catalogID := c.Param("catalog_id")
	catalog, ok := s.Catalogs[catalogID]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, catalog)
}

func (s *Scenario) postActivate(c echo.Context) error {
	state := &IsuConditionPosterRequest{}
	err := c.Bind(state)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if !(0 <= state.TargetPort && state.TargetPort < 0x1000) {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	targetBaseURL := fmt.Sprintf(
		"http://%s:%d",
		state.TargetIP, state.TargetPort,
	)

	//poster Goroutineの起動
	var isuDetail *IsuDetailInfomation
	var scenarioChan *model.StreamsForPoster
	posterContext, cancelFunc := context.WithCancel(jiaAPIContext)
	err = func() error {
		var ok bool
		streamsForPosterMutex.Lock()
		defer streamsForPosterMutex.Unlock()
		scenarioChan, ok = streamsForPoster[state.IsuUUID]
		if !ok {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		v, ok := isuIsActivated[state.IsuUUID]
		if ok && v.activated {
			return echo.NewHTTPError(http.StatusForbidden)
		}
		isuIsActivated[state.IsuUUID] = JiaAPI2PosterData{
			activated:  true,
			cancelFunc: cancelFunc,
		}
		isuDetail = isuDetailInfomation[state.IsuUUID]

		return nil
	}()
	if err != nil {
		return err
	}
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		s.keepPosting(posterContext, jiaAPIStep, targetBaseURL, state.IsuUUID, scenarioChan)
	}()

	return c.JSON(http.StatusAccepted, isuDetail)
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
	v.cancelFunc()
	v.activated = false

	return c.NoContent(http.StatusNoContent)
}
