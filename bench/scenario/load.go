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
	// ctx, cancel := context.WithTimeout(parent, 60*time.Second) //mainで指定している方を見るべき
	// defer cancel()
	ctx := parent

	logger.ContestantLogger.Printf("===> LOAD")
	logger.AdminLogger.Printf("LOAD INFO\n  Language: %s\n  Campaign: None\n", s.Language)

	/*
		TODO: 実際の負荷走行シナリオ
	*/

	//通常ユーザー
	s.AddNormalUser(ctx, step, 10)
	//マニアユーザー
	s.AddManiacUser(ctx, step, 2)
	//企業ユーザー
	s.AddCompanyUser(ctx, step, 1)

	<-ctx.Done()
	s.jiaChancel()
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
	defer logger.AdminLogger.Println("Normal User END")

	//ユーザー作成
	userAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	user := s.NewUser(ctx, step, userAgent, model.UserTypeNormal)
	if user == nil {
		logger.AdminLogger.Println("Normal User fail: NewUser")
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
		isu := s.NewIsu(ctx, step, user, true)
		if isu == nil {
			logger.AdminLogger.Println("Normal User fail: NewIsu(initialize)")
			return //致命的でないエラー
		}
	}
	step.AddScore(ScoreNormalUserInitialize)

	randEngine := rand.New(rand.NewSource(5498513))
	nextTargetIsuIndex := 0
	scenarioDoneCount := 0
	scenarioSuccess := false
	lastSolvedTime := s.virtualTimeStart
scenarioLoop:
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(500 * time.Millisecond) //TODO: 頻度調整
		if scenarioSuccess {
			scenarioDoneCount++
			step.AddScore(ScoreNormalUserLoop) //TODO: 得点条件の修正

			//シナリオに成功している場合は椅子追加
			if isuCount < scenarioDoneCount/30 && isuCount < isuCountMax {
				isu := s.NewIsu(ctx, step, user, true)
				if isu == nil {
					logger.AdminLogger.Println("Normal User fail: NewIsu")
				} else {
					isuCount++
				}
				//logger.AdminLogger.Printf("Normal User Isu: %d\n", isuCount)
			}
		}
		scenarioSuccess = true

		//posterからconditionの取得
		for _, isu := range user.IsuListOrderByCreatedAt {
		getConditionFromPosterLoop:
			for {
				select {
				case <-ctx.Done():
					return
				case conditions, ok := <-isu.StreamsForScenario.ConditionChan:
					if !ok {
						break getConditionFromPosterLoop
					}
					for _, c := range conditions {
						isu.Conditions.Add(&c)
					}
				default:
					break getConditionFromPosterLoop
				}
			}
		}

		//TODO: 乱数にする
		nextTargetIsuIndex += 1
		nextTargetIsuIndex %= isuCount
		targetIsu := user.IsuListOrderByCreatedAt[nextTargetIsuIndex]

		//GET /
		realNow := time.Now()
		virtualNow := s.ToVirtualTime(realNow)
		_, _, errs := browserGetHomeAction(ctx, user.Agent, virtualNow.Unix(),
			func(res *http.Response, isuList []*service.Isu) []error {
				return verifyIsuOrderByCreatedAt(res, user.IsuListOrderByCreatedAt, isuList)
			},
			func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
				//TODO: conditionの検証
				return []error{}
			},
		)
		for _, err := range errs {
			scenarioSuccess = false
			step.AddError(err)
		}

		//GET /isu/{jia_isu_uuid}
		_, _, errs = browserGetIsuDetailAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
			func(res *http.Response, catalog *service.Catalog) []error {
				//TODO: catalogの検証
				//targetIsu.JIACatalogID
				//return verifyCatalog(res, , catalog)
				return []error{}
			},
		)
		for _, err := range errs {
			scenarioSuccess = false
			step.AddError(err)
		}

		if randEngine.Intn(3) < 2 {
			//TODO: リロード

			//定期的にconditionを見に行くシナリオ
			realNow = time.Now()
			virtualNow = s.ToVirtualTime(realNow)
			_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
				service.GetIsuConditionRequest{
					StartTime:        nil,
					CursorEndTime:    uint64(virtualNow.Unix()),
					CursorJIAIsuUUID: "",
					ConditionLevel:   "info,warning,critical",
					Limit:            nil,
				},
				func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
					//conditionの検証
					err := verifyIsuConditions(res, &targetIsu.Conditions, model.ConditionLevelInfo|model.ConditionLevelWarning|model.ConditionLevelCritical,
						model.IsuConditionCursor{TimestampUnix: virtualNow.Unix(), OwnerID: targetIsu.JIAIsuUUID}, targetIsu.Owner.IsuListByID,
						conditions, s.ToVirtualTime(realNow.Add(-1*time.Second)).Unix(),
					)
					if err != nil {
						return []error{err}
					}
					return []error{}
				},
			)
			for _, err := range errs {
				scenarioSuccess = false
				step.AddError(err)
			}
			if len(errs) > 0 || len(conditions) == 0 {
				continue scenarioLoop
			}

			//スクロール
			for i := 0; i < 2 && len(conditions) == 20*(i+1); i++ {
				var conditionsTmp []*service.GetIsuConditionResponse
				CursorEndTime := conditions[len(conditions)-1].Timestamp
				conditionsTmp, res, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
					service.GetIsuConditionRequest{
						StartTime:        nil,
						CursorEndTime:    uint64(CursorEndTime),
						CursorJIAIsuUUID: "",
						ConditionLevel:   "info,warning,critical",
						Limit:            nil,
					},
				)
				if err != nil {
					scenarioSuccess = false
					step.AddError(err)
					break
				}
				//検証
				err = verifyIsuConditions(res, &targetIsu.Conditions, model.ConditionLevelInfo|model.ConditionLevelWarning|model.ConditionLevelCritical,
					model.IsuConditionCursor{TimestampUnix: CursorEndTime, OwnerID: targetIsu.JIAIsuUUID}, targetIsu.Owner.IsuListByID,
					conditions, s.ToVirtualTime(realNow.Add(-1*time.Second)).Unix(),
				)
				if err != nil {
					scenarioSuccess = false
					step.AddError(err)
					break
				}

				conditions = append(conditions, conditionsTmp...)
			}

			//conditionを確認して、椅子状態を改善
			solvedCondition, findTimestamp := findBadIsuState(conditions)
			if solvedCondition != model.IsuStateChangeNone && lastSolvedTime.Before(time.Unix(findTimestamp, 0)) {
				//graphを見る
				virtualDay := (findTimestamp / (24 * 60 * 60)) * (24 * 60 * 60)
				_, _, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualDay),
					func(res *http.Response, graph []*service.GraphResponse) []error {
						return []error{} //TODO: 検証
					},
				)
				for _, err := range errs {
					scenarioSuccess = false
					step.AddError(err)
				}

				//状態改善
				lastSolvedTime = time.Unix(findTimestamp, 0)
				targetIsu.StreamsForScenario.StateChan <- solvedCondition //バッファがあるのでブロック率は低い読みで直列に投げる
			}
		} else {

			//TODO: graphを見に行くシナリオ
			realNow = time.Now()
			virtualNow = s.ToVirtualTime(realNow)
			virtualToday := time.Date(virtualNow.Year(), virtualNow.Month(), virtualNow.Day(), 0, 0, 0, 0, virtualNow.Location())
			_, graphToday, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualToday.Unix()),
				func(res *http.Response, graph []*service.GraphResponse) []error {
					return []error{} //TODO: 検証
				},
			)
			for _, err := range errs {
				scenarioSuccess = false
				step.AddError(err)
			}
			if len(errs) > 0 {
				continue scenarioLoop
			}

			//前日のグラフ
			_, _, errs = browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualToday.Add(-24*time.Hour).Unix()),
				func(res *http.Response, graph []*service.GraphResponse) []error {
					return []error{} //TODO: 検証
				},
			)
			for _, err := range errs {
				scenarioSuccess = false
				step.AddError(err)
			}
			if len(errs) > 0 {
				continue scenarioLoop
			}

			//悪いものを探す
			var errorEndAtUnix int64 = 0
			for _, g := range graphToday {
				if g.Data != nil && g.Data.Score < 100 {
					errorEndAtUnix = g.StartAt
				}
			}

			//悪いものがあれば、そのconditionを取る
			if errorEndAtUnix != 0 {
				startTime := uint64(errorEndAtUnix - 60*60)
				_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
					service.GetIsuConditionRequest{
						StartTime:        &startTime,
						CursorEndTime:    uint64(errorEndAtUnix),
						CursorJIAIsuUUID: "",
						ConditionLevel:   "warning,critical",
						Limit:            nil,
					},
					func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
						err := verifyIsuConditions(res, &targetIsu.Conditions, model.ConditionLevelWarning|model.ConditionLevelCritical,
							model.IsuConditionCursor{TimestampUnix: errorEndAtUnix, OwnerID: targetIsu.JIAIsuUUID}, targetIsu.Owner.IsuListByID,
							conditions, s.ToVirtualTime(realNow.Add(-1*time.Second)).Unix(),
						)
						if err != nil {
							return []error{err}
						}
						return []error{}
					},
				)
				for _, err := range errs {
					scenarioSuccess = false
					step.AddError(err)
				}
				if len(errs) > 0 {
					continue scenarioLoop
				}

				//状態改善
				solvedCondition, findTimestamp := findBadIsuState(conditions)
				if solvedCondition != model.IsuStateChangeNone && lastSolvedTime.Before(time.Unix(findTimestamp, 0)) {
					lastSolvedTime = time.Unix(findTimestamp, 0)
					targetIsu.StreamsForScenario.StateChan <- solvedCondition //バッファがあるのでブロック率は低い読みで直列に投げる
				}
			}
		}
	}
}

func findBadIsuState(conditions []*service.GetIsuConditionResponse) (model.IsuStateChange, int64) {
	//TODO: すでに改善済みのものを弾く

	var virtualTimestamp int64
	solvedCondition := model.IsuStateChangeNone
	for _, c := range conditions {
		//MEMO: 重かったらフォーマットが想定通りの前提で最適化する
		bad := false
		for _, cond := range strings.Split(c.Condition, ",") {
			keyValue := strings.Split(cond, "=")
			if len(keyValue) != 2 {
				continue //形式に従っていないものは無視
			}
			if keyValue[1] != "false" {
				bad = true
				if keyValue[0] == "is_dirty" {
					solvedCondition |= model.IsuStateChangeClear
				} else if keyValue[0] == "is_overweight" {
					solvedCondition |= model.IsuStateChangeDetectOverweight
				} else if keyValue[0] == "is_broken" {
					solvedCondition |= model.IsuStateChangeRepair
				}
			}
		}
		if bad && virtualTimestamp == 0 {
			virtualTimestamp = c.Timestamp
		}
	}

	return solvedCondition, virtualTimestamp
}
