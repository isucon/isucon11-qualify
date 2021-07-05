package scenario

// load.go
// シナリオの内、loadフェーズの処理

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/service"
)

func (s *Scenario) Load(parent context.Context, step *isucandar.BenchmarkStep) error {
	if s.NoLoad {
		return nil
	}
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()

	logger.ContestantLogger.Printf("===> LOAD")
	logger.AdminLogger.Printf("LOAD INFO\n  Language: %s\n  Campaign: None\n", s.Language)

	/*
		TODO: 実際の負荷走行シナリオ
	*/

	//通常ユーザー
	normalUserWorker, err := worker.NewWorker(func(ctx context.Context, _ int) {
		defer s.loadWaitGroup.Done()
		s.loadNormalUser(ctx, step)
	}, worker.WithInfinityLoop())
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	s.normalUserWorker = normalUserWorker
	//マニアユーザー
	maniacUserWorker, err := worker.NewWorker(func(ctx context.Context, _ int) {
		defer s.loadWaitGroup.Done()
		//s.loadManiacUser(ctx, step) //TODO:
	}, worker.WithInfinityLoop())
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	s.maniacUserWorker = maniacUserWorker
	//企業ユーザー
	companyUserWorker, err := worker.NewWorker(func(ctx context.Context, _ int) {
		defer s.loadWaitGroup.Done()
		//s.loadCompanyUser(ctx, step) //TODO:
	}, worker.WithInfinityLoop())
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	s.companyUserWorker = companyUserWorker

	<-ctx.Done()
	s.loadWaitGroup.Wait()

	return nil
}

func (s *Scenario) loadNormalUser(ctx context.Context, step *isucandar.BenchmarkStep) {

	select {
	case <-ctx.Done():
		return
	default:
	}

	//ユーザー作成
	userAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	user := s.NewUser(ctx, step, userAgent)
	if user == nil {
		return //致命的でないエラー
	}
	func() {
		s.normalUsersMtx.Lock()
		defer s.normalUsersMtx.Unlock()
		s.normalUsers = append(s.normalUsers, user)
	}()

	//椅子作成
	const isuCountMax = 4 //ルートページに表示する最大数
	isuCount := 1
	for i := 0; i < isuCount; i++ {
		_ = s.NewIsu(ctx, step, user, true)
	}

	randEngine := rand.New(rand.NewSource(5498513))
	scenarioDoneCount := 0
	nextTargetIsuIndex := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		//TODO: 乱数にする
		nextTargetIsuIndex += 1
		nextTargetIsuIndex %= isuCount
		targetIsu := user.IsuListOrderByCreatedAt[nextTargetIsuIndex]

		//GET /
		_, _, errs := browserGetHomeAction(ctx, user.Agent,
			func(res *http.Response, isuList []*service.Isu) []error {
				return verifyIsuOrderByCreatedAt(res, user.IsuListOrderByCreatedAt, isuList)
			},
			func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
				//TODO: conditionの検証
				return []error{}
			},
		)
		for _, err := range errs {
			step.AddError(err)
		}

		//GET /isu/{jia_isu_uuid}
		browserGetIsuDetailAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
			func(res *http.Response, catalog *service.Catalog) []error {
				//TODO: catalogの検証
				//targetIsu.JIACatalogID
				//return verifyCatalog(res, , catalog)
				return []error{}
			},
		)

		if randEngine.Intn(3) < 2 {
			//定期的にconditionを見に行くシナリオ

		} else {

			//TODO: graphを見に行くシナリオ
		}

		scenarioDoneCount++
		//TODO: 椅子の追加
	}
}
