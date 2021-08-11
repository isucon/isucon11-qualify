package scenario

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	streamsForPosterMutex sync.Mutex
	isuIsActivated        = map[string]struct{}{}
	streamsForPoster      = map[string]*model.StreamsForPoster{}
	//isuDetailInfomation   = map[string]*IsuDetailInfomation{}
	isuFromUUID = map[string]*model.Isu{}

	jiaAPIContext context.Context
)

type IsuDetailInfomation struct {
	Character string `json:"character"`
}

//シナリオ Goroutineからの呼び出し
func RegisterToJiaAPI(isu *model.Isu, streams *model.StreamsForPoster) {
	streamsForPosterMutex.Lock()
	defer streamsForPosterMutex.Unlock()
	isuFromUUID[isu.JIAIsuUUID] = isu
	streamsForPoster[isu.JIAIsuUUID] = streams
}

func (s *Scenario) JiaAPIService(ctx context.Context) {
	defer logger.AdminLogger.Println("--- JiaAPIService END")

	jiaAPIContext = ctx

	// Echo instance
	e := echo.New()
	e.HideBanner = true
	//e.Debug = true
	//e.Logger.SetLevel(log.DEBUG)

	// Middleware
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.POST("/api/activate", func(c echo.Context) error { return s.postActivate(c) })

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
	state := &service.JIAServiceRequest{}
	err := c.Bind(state)
	if err != nil {
		return c.String(http.StatusBadRequest, "Bad Request")
	}
	targetBaseURL, err := url.Parse(state.TargetBaseURL)
	if err != nil {
		return c.String(http.StatusBadRequest, "Bad URL")
	}
	//TODO: URLの検証

	//poster Goroutineの起動
	var isu *model.Isu
	var scenarioChan *model.StreamsForPoster
	posterContext := jiaAPIContext
	errCode, errMsg := func() (int, string) {
		var ok bool
		streamsForPosterMutex.Lock()
		defer streamsForPosterMutex.Unlock()
		// scenario goroutine とやり取りするためのチャネルを受け取る
		scenarioChan, ok = streamsForPoster[state.IsuUUID]
		if !ok {
			return http.StatusNotFound, "Bad isu_uuid"
		}
		isu, ok = isuFromUUID[state.IsuUUID]
		if !ok {
			//scenarioChanでチェックしているのでここには来ないはず
			return http.StatusNotFound, "Bad isu_uuid"
		}
		_, ok = isuIsActivated[state.IsuUUID]
		if ok {
			//activate済み
			return 0, ""
		}

		// activate 済みフラグを立てる
		isuIsActivated[state.IsuUUID] = struct{}{}

		//activate
		s.loadWaitGroup.Add(1)
		go func() {
			defer s.loadWaitGroup.Done()
			s.keepPosting(posterContext, targetBaseURL, isu, scenarioChan)
		}()
		return 0, ""
	}()
	if errCode != 0 {
		return c.String(errCode, errMsg)
	}

	time.Sleep(50 * time.Millisecond)
	return c.JSON(http.StatusAccepted, IsuDetailInfomation{isu.Character})
}
