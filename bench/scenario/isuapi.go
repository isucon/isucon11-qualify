package scenario

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	isuPosterMutex sync.Mutex
	isuPosterData  = map[string]*model.IsuPosterChan{}
)

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
			step.AddError(failure.NewError(ErrCritical, fmt.Errorf("ISU協会サービスが異常終了しました")))
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
	//TODO:
	return fmt.Errorf("NotImplemented")
}

func postDeactivate(c echo.Context) error {
	//TODO:
	return fmt.Errorf("NotImplemented")
}
