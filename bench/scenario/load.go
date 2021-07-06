package scenario

// load.go
// シナリオの内、loadフェーズの処理

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

func (s *Scenario) Load(parent context.Context, step *isucandar.BenchmarkStep) error {
	defer s.jiaChancel()
	step.Result().Score.Reset()
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
	}, worker.WithInfinityLoop(), worker.WithMaxParallelism(10))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	s.normalUserWorker = normalUserWorker
	go s.normalUserWorker.Process(ctx)
	//マニアユーザー
	maniacUserWorker, err := worker.NewWorker(func(ctx context.Context, _ int) {
		defer s.loadWaitGroup.Done()
		<-ctx.Done() //TODO: 消す
		//s.loadManiacUser(ctx, step) //TODO:
	}, worker.WithInfinityLoop(), worker.WithMaxParallelism(1))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	s.maniacUserWorker = maniacUserWorker
	go s.maniacUserWorker.Process(ctx)
	//企業ユーザー
	companyUserWorker, err := worker.NewWorker(func(ctx context.Context, _ int) {
		defer s.loadWaitGroup.Done()
		<-ctx.Done() //TODO: 消す
		//s.loadCompanyUser(ctx, step) //TODO:
	}, worker.WithInfinityLoop(), worker.WithMaxParallelism(1))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	s.companyUserWorker = companyUserWorker
	go s.companyUserWorker.Process(ctx)

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
	logger.AdminLogger.Println("Normal User start")

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
	nextTargetIsuIndex := 0
	for scenarioDoneCount := 0; true; scenarioDoneCount++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		//posterからconditionの取得
		for _, isu := range user.IsuListOrderByCreatedAt {
		getConditionFromPosterLoop:
			for {
				select {
				case <-ctx.Done():
					return
				case cond, ok := <-isu.StreamsForScenario.ConditionChan:
					if !ok {
						break getConditionFromPosterLoop
					}
					isu.Conditions = append(isu.Conditions, *cond)
				default:
				}
			}
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
			//TODO: リロード

			//定期的にconditionを見に行くシナリオ
			virtualNow := s.ToVirtualTime(time.Now())
			_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
				service.GetIsuConditionRequest{
					StartTime:        nil,
					CursorEndTime:    uint64(virtualNow.Unix()),
					CursorJIAIsuUUID: "",
					ConditionLevel:   "info,warning,critical",
					Limit:            nil,
				},
				func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
					return []error{}
				},
			)
			for _, err := range errs {
				step.AddError(err)
			}
			if len(errs) > 0 {
				continue
			}

			//スクロール
			var res *http.Response
			for i := 0; i < 2 && len(conditions) == 20*(i+1); i++ {
				var conditionsTmp []*service.GetIsuConditionResponse
				conditionsTmp, res, err = getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
					service.GetIsuConditionRequest{
						StartTime:        nil,
						CursorEndTime:    uint64(conditions[len(conditions)-1].Timestamp),
						CursorJIAIsuUUID: "",
						ConditionLevel:   "info,warning,critical",
						Limit:            nil,
					},
				)
				if err != nil {
					step.AddError(err)
					break
				} else {
					conditions = append(conditions, conditionsTmp...)
				}
			}

			//TODO: conditionの検証
			if res != nil { //エラーつぶし
			}

			//conditionを確認して、椅子状態を改善
			solvedCondition := model.IsuStateChangeNone
			for _, c := range conditions {
				//MEMO: 重かったらフォーマットが想定通りの前提で最適化する
				for _, cond := range strings.Split(c.Condition, ",") {
					keyValue := strings.Split(cond, "=")
					if len(keyValue) != 2 {
						continue //形式に従っていないものは無視
					}
					if keyValue[1] != "false" {
						if keyValue[0] == "is_dirty" {
							solvedCondition |= model.IsuStateChangeClear
						} else if keyValue[0] == "is_overweight" {
							solvedCondition |= model.IsuStateChangeDetectOverweight
						} else if keyValue[0] == "is_broken" {
							solvedCondition |= model.IsuStateChangeRepair
						}
					}
				}
			}

			if solvedCondition != model.IsuStateChangeNone {
				//TODO: graph

				targetIsu.StreamsForScenario.StateChan <- solvedCondition
			}
		} else {

			//TODO: graphを見に行くシナリオ
		}

		//TODO: 椅子の追加
	}
}
