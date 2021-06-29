package scenario

// prepare.go
// シナリオの内、prepareフェーズの処理

import (
	"context"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	logger.ContestantLogger.Printf("===> PREPARE")

	//initialize
	initializer, err := s.NewAgent(
		agent.WithNoCache(), agent.WithNoCookie(), agent.WithTimeout(20*time.Second),
	)
	if err != nil {
		return failure.NewError(ErrCritical, err)
	}
	initializer.Name = "benchmarker-initializer"

	initResponse, errs := initializeAction(ctx, initializer)
	for _, err := range errs {
		step.AddError(err)
	}
	if len(errs) > 0 {
		//return ErrScenarioCancel
		return ErrCritical
	}

	s.Language = initResponse.Language

	//各エンドポイントのチェック
	err = s.prepareCheckAuth(ctx, step)
	if err != nil {
		return err
	}

	return nil
}

//エンドポイント事の単体テスト

func (s *Scenario) prepareCheckAuth(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {

		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		switch index % 10 {
		default:
			//ログイン成功
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
			}
		case 1:
			//Unexpected signing method, StatusForbidden
			jwtHS256, err := service.GenerateHS256JWT(userID, time.Now())
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithForbiddenJWT(ctx, agt, jwtHS256)
			for _, err := range errs {
				step.AddError(err)
			}
		case 2:
			//expired, StatusForbidden
			jwtExpired, err := service.GenerateJWT(userID, time.Now().Add(-365*24*time.Hour))
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithForbiddenJWT(ctx, agt, jwtExpired)
			for _, err := range errs {
				step.AddError(err)
			}
		case 3:
			//jwt is missing, StatusForbidden
			errs := authActionWithoutJWT(ctx, agt)
			for _, err := range errs {
				step.AddError(err)
			}
		case 4:
			//invalid private key, StatusForbidden
			jwtDummyKey, err := service.GenerateDummyJWT(userID, time.Now())
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithForbiddenJWT(ctx, agt, jwtDummyKey)
			for _, err := range errs {
				step.AddError(err)
			}
		case 5:
			//not jwt, StatusForbidden
			errs := authActionWithForbiddenJWT(ctx, agt, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.")
			for _, err := range errs {
				step.AddError(err)
			}
		case 6:
			//偽装されたjwt, StatusForbidden
			userID2, err := model.MakeRandomUserID()
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			jwtTampered, err := service.GenerateTamperedJWT(userID, userID2, time.Now())
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithForbiddenJWT(ctx, agt, jwtTampered)
			for _, err := range errs {
				step.AddError(err)
			}
		case 7:
			//jia_user_id is missing, StatusBadRequest
			jwtNoData, err := service.GenerateJWTWithNoData(time.Now())
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithInvalidJWT(ctx, agt, jwtNoData, http.StatusBadRequest, "invalid JWT payload")
			for _, err := range errs {
				step.AddError(err)
			}
		case 8:
			//jwt with invalid data type, StatusBadRequest
			jwtInvalidDataType, err := service.GenerateJWTWithInvalidType(userID, time.Now())
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithInvalidJWT(ctx, agt, jwtInvalidDataType, http.StatusBadRequest, "invalid JWT payload")
			for _, err := range errs {
				step.AddError(err)
			}
		}

	}, worker.WithLoopCount(20))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	//w.Wait() //念のためもう一度止まってるか確認

	//作成済みユーザーへのログイン確認
	agt, err := s.NewAgent()
	if err != nil {
		step.AddError(failure.NewError(ErrCritical, err))
		return nil
	}
	userID, err := model.MakeRandomUserID()
	if err != nil {
		step.AddError(failure.NewError(ErrCritical, err))
		return nil
	}

	_, errs := authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
	}
	agt.ClearCookie()
	//二回目のログイン
	_, errs = authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
	}

	return nil
}
