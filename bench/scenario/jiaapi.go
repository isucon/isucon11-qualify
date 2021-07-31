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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	streamsForPosterMutex sync.Mutex
	isuIsActivated        = map[string]JiaAPI2PosterData{}
	streamsForPoster      = map[string]*model.StreamsForPoster{}
	//isuDetailInfomation   = map[string]*IsuDetailInfomation{}
	isuFromUUID      = map[string]*model.Isu{}
	isuTargetBaseUrl = map[string]string{} // 本当はISUに紐付けたい

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
	closeWait  <-chan struct{}
	cancelFunc context.CancelFunc
}

//シナリオ Goroutineからの呼び出し
func RegisterToJiaAPI(isu *model.Isu, streams *model.StreamsForPoster) {
	streamsForPosterMutex.Lock()
	defer streamsForPosterMutex.Unlock()
	isuFromUUID[isu.JIAIsuUUID] = isu
	streamsForPoster[isu.JIAIsuUUID] = streams
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
	e.POST("/api/activate", func(c echo.Context) error { return s.postActivate(c) })
	e.POST("/api/deactivate", postDeactivate)

	// Start
	var bindPort string
	if s.jiaServiceURL.Port() != "" {
		bindPort = fmt.Sprintf("0.0.0.0:%s", s.jiaServiceURL.Port())
	} else {
		bindPort = "0.0.0.0:80"
	}
	s.loadWaitGroup.Add(1)
	go func() {
		defer logger.AdminLogger.Println("--- ISU協会サービス END")
		defer s.loadWaitGroup.Done()
		err := e.Start(bindPort)
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
	var isu *model.Isu
	var scenarioChan *model.StreamsForPoster
	closeWait := make(chan struct{})
	posterContext, cancelFunc := context.WithCancel(jiaAPIContext)
	err = func() error {
		var ok bool
		streamsForPosterMutex.Lock()
		defer streamsForPosterMutex.Unlock()
		// scenario goroutine とやり取りするためのチャネルを受け取る
		scenarioChan, ok = streamsForPoster[state.IsuUUID]
		if !ok {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		// activate 済みであれば 403 を返す
		v, ok := isuIsActivated[state.IsuUUID]
		if ok && v.activated {
			return echo.NewHTTPError(http.StatusForbidden)
		}
		// activate 済みフラグを立てる
		isuIsActivated[state.IsuUUID] = JiaAPI2PosterData{
			activated:  true,
			cancelFunc: cancelFunc,
			closeWait:  closeWait,
		}
		isuTargetBaseUrl[state.IsuUUID] = targetBaseURL
		// リクエストされた JIA_ISU_UUID が事前に scenario.NewIsu にて作成された isu と紐付かない場合 403 を返す
		isu, ok = isuFromUUID[state.IsuUUID]
		if !ok {
			return echo.NewHTTPError(http.StatusForbidden)
		}

		return nil
	}()
	if err != nil {
		return err
	}
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		s.keepPosting(posterContext, jiaAPIStep, targetBaseURL, isu, scenarioChan, closeWait)
	}()

	time.Sleep(50 * time.Millisecond)
	return c.JSON(http.StatusAccepted, IsuDetailInfomation{isu.JIACatalogID, isu.Character})
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
	<-v.closeWait //posterの終了を待機

	time.Sleep(50 * time.Millisecond)
	return c.NoContent(http.StatusNoContent)
}
