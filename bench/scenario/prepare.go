package scenario

// prepare.go
// シナリオの内、prepareフェーズの処理

import (
	"context"
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

	return nil
}

//エンドポイント事の単体テスト

func (s *Scenario) prepareCheckAuth(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {

		a, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		switch index {
		default:
			//ログイン成功
			_, errs := authAction(ctx, a, userID)
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
			errs := authActionWithForbiddenJWT(ctx, a, jwtHS256)
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
			errs := authActionWithForbiddenJWT(ctx, a, jwtExpired)
			for _, err := range errs {
				step.AddError(err)
			}
		case 3:
			//jwt is missing, StatusForbidden
			//TODO:
		case 4:
			//invalid private key, StatusForbidden
			jwtDummyKey, err := service.GenerateDummyJWT(userID, time.Now())
			if err != nil {
				step.AddError(failure.NewError(ErrCritical, err))
				return
			}
			errs := authActionWithForbiddenJWT(ctx, a, jwtDummyKey)
			for _, err := range errs {
				step.AddError(err)
			}
		case 5:
			//jia_user_id is missing, StatusBadRequest
			//TODO:
		case 6:
			//not jwt, StatusForbidden
			errs := authActionWithForbiddenJWT(ctx, a, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.")
			for _, err := range errs {
				step.AddError(err)
			}
		case 7:
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
			errs := authActionWithForbiddenJWT(ctx, a, jwtTampered)
			for _, err := range errs {
				step.AddError(err)
			}
		}

	}, worker.WithLoopCount(10))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	//w.Wait() //念のためもう一度止まってるか確認

	//作成済みユーザーへのログイン確認
	//TODO:

	return nil
}
