package scenario

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

	posterRootContext context.Context
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

	posterRootContext, s.JiaPosterCancel = context.WithCancel(ctx)

	// Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
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
	go func() {
		defer logger.AdminLogger.Println("--- ISU協会サービス END")
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

	//poster Goroutineの起動
	var isu *model.Isu
	var scenarioChan *model.StreamsForPoster
	var fqdn string
	posterContext := posterRootContext
	errCode, errMsg := func() (int, string) {
		var ok bool
		streamsForPosterMutex.Lock()
		defer streamsForPosterMutex.Unlock()
		// scenario goroutine とやり取りするためのチャネルを受け取る
		scenarioChan, ok = streamsForPoster[state.IsuUUID]
		if !ok {
			return http.StatusNotFound, "Bad isu_uuid"
		}
		// リクエストされた JIA_ISU_UUID が事前に scenario.NewIsu にて作成された isu と紐付かない場合 404 を返す
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

		// useTLS が有効 && POST isucondition する URL に https 以外が指定されていたら 400 を返す
		if s.UseTLS && targetBaseURL.Scheme != "https" {
			return http.StatusBadRequest, "Bad URL Scheme: scheme must be https"
		}
		// FQDN が競技者 VM のものでない場合 400 を返す
		fqdn = targetBaseURL.Hostname()
		ipAddr, ok := s.GetIPAddrFromFqdn(fqdn)
		if !ok {
			return http.StatusBadRequest, "Bad URL: hostname must be isucondition-[1-3].t.isucon.dev"
		}
		//httpsモードの際はportは指定なしのみ
		port := targetBaseURL.Port()
		if s.UseTLS && port != "" {
			return http.StatusBadRequest, "Bad Port: ポート番号は指定できません"
		}
		// URL の文字列を IP アドレスに変換
		if port != "" {
			targetBaseURL.Host = strings.Join([]string{ipAddr, port}, ":")
		} else {
			targetBaseURL.Host = ipAddr
		}

		// activate 済みフラグを立てる
		isuIsActivated[state.IsuUUID] = struct{}{}
		//activate
		s.loadWaitGroup.Add(1)
		go func() {
			defer s.loadWaitGroup.Done()
			defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() keepPosting")
			s.keepPosting(posterContext, targetBaseURL, fqdn, isu, scenarioChan)
		}()
		return 0, ""
	}()
	if errCode != 0 {
		return c.String(errCode, errMsg)
	}

	time.Sleep(50 * time.Millisecond)
	return c.JSON(http.StatusAccepted, IsuDetailInfomation{isu.Character})
}
