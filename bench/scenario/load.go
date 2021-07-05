package scenario

// load.go
// シナリオの内、loadフェーズの処理

import (
	"context"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon11-qualify/bench/logger"
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
	isuCount := 3
	for i := 0; i < isuCount; i++ {
		_ = s.NewIsu(ctx, step, user, true)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		//TODO:
	}
}
