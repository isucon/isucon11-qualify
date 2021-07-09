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
		user.GetConditionFromChan(ctx)
		select {
		case <-ctx.Done():
			return
		default:
		}

		//TODO: 乱数にする
		nextTargetIsuIndex += 1
		nextTargetIsuIndex %= isuCount
		targetIsu := user.IsuListOrderByCreatedAt[nextTargetIsuIndex]
		mustExistUntil := s.ToVirtualTime(time.Now().Add(-1 * time.Second)).Unix()

		//GET /
		dataExistTimestamp := GetConditionDataExistTimestamp(s, user)
		_, _, errs := browserGetHomeAction(ctx, user.Agent, dataExistTimestamp,
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
			request := service.GetIsuConditionRequest{
				StartTime:        nil,
				CursorEndTime:    uint64(dataExistTimestamp),
				CursorJIAIsuUUID: "",
				ConditionLevel:   "info,warning,critical",
				Limit:            nil,
			}
			_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
				request,
				func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
					//conditionの検証
					err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request,
						conditions, mustExistUntil,
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
				request = service.GetIsuConditionRequest{
					StartTime:        nil,
					CursorEndTime:    uint64(CursorEndTime),
					CursorJIAIsuUUID: "",
					ConditionLevel:   "info,warning,critical",
					Limit:            nil,
				}
				conditionsTmp, res, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
				if err != nil {
					scenarioSuccess = false
					step.AddError(err)
					break
				}
				//検証
				//ここは、古いデータのはずなのでconditionのchanからの再取得は要らない
				err = verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request,
					conditionsTmp, mustExistUntil,
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
			virtualToday := (dataExistTimestamp / (24 * 60 * 60)) * (24 * 60 * 60)
			_, graphToday, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualToday),
				func(res *http.Response, graph []*service.GraphResponse) []error {
					//検証前にデータ取得
					user.GetConditionFromChan(ctx)
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
			_, _, errs = browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualToday-60*60),
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
				request := service.GetIsuConditionRequest{
					StartTime:        &startTime,
					CursorEndTime:    uint64(errorEndAtUnix),
					CursorJIAIsuUUID: "",
					ConditionLevel:   "warning,critical",
					Limit:            nil,
				}
				_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
					request,
					func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
						//検証
						//ここは、古いデータのはずなのでconditionのchanからの再取得は要らない
						//TODO: starttimeの検証
						err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request,
							conditions, mustExistUntil,
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

func (s *Scenario) loadCompanyUser(ctx context.Context, step *isucandar.BenchmarkStep) {

	select {
	case <-ctx.Done():
		return
	default:
	}
	logger.AdminLogger.Println("Company User start")
	defer logger.AdminLogger.Println("Company User END")

	//ユーザー作成
	userAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	user := s.NewUser(ctx, step, userAgent, model.UserTypeCompany)
	if user == nil {
		logger.AdminLogger.Println("Company User fail: NewUser")
		return //致命的でないエラー
	}
	func() {
		s.companyUsersMtx.Lock()
		defer s.companyUsersMtx.Unlock()
		s.companyUsers = append(s.companyUsers, user)
	}()

	//椅子作成
	const isuCountMax = 1000
	isuCount := 50
	for i := 0; i < isuCount; i++ {
		isu := s.NewIsu(ctx, step, user, true)
		if isu == nil {
			logger.AdminLogger.Println("Company User fail: NewIsu(initialize)")
			return //致命的でないエラー
		}
	}
	step.AddScore(ScoreCompanyUserInitialize)

	randEngine := rand.New(rand.NewSource(5498513))
	nextTargetIsuIndex := 0
	scenarioDoneCount := 0
	scenarioSuccess := false
	lastSolvedTime := s.virtualTimeStart
scenarioLoop:
	for {
		//TODO: 今はnormal userそのままになっているので、ちゃんと企業ユーザー用に書き直す

		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(500 * time.Millisecond) //TODO: 頻度調整
		if scenarioSuccess {
			scenarioDoneCount++
			step.AddScore(ScoreCompanyUserLoop) //TODO: 得点条件の修正

			//シナリオに成功している場合は椅子追加
			for isuCount < (scenarioDoneCount/30)*50 && isuCount < isuCountMax {
				isu := s.NewIsu(ctx, step, user, true)
				if isu == nil {
					logger.AdminLogger.Println("Company User fail: NewIsu")
				} else {
					isuCount++
				}
				//logger.AdminLogger.Printf("Company User Isu: %d\n", isuCount)
			}
		}
		scenarioSuccess = true

		//posterからconditionの取得
		user.GetConditionFromChan(ctx)
		select {
		case <-ctx.Done():
			return
		default:
		}

		//TODO: 乱数にする
		nextTargetIsuIndex += 1
		nextTargetIsuIndex %= isuCount
		targetIsu := user.IsuListOrderByCreatedAt[nextTargetIsuIndex]
		mustExistUntil := s.ToVirtualTime(time.Now().Add(-1 * time.Second)).Unix()

		//GET /
		dataExistTimestamp := GetConditionDataExistTimestamp(s, user)
		_, _, errs := browserGetHomeAction(ctx, user.Agent, dataExistTimestamp,
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
			request := service.GetIsuConditionRequest{
				StartTime:        nil,
				CursorEndTime:    uint64(dataExistTimestamp),
				CursorJIAIsuUUID: "",
				ConditionLevel:   "info,warning,critical",
				Limit:            nil,
			}
			_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
				request,
				func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
					//conditionの検証
					err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request,
						conditions, mustExistUntil,
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
				request = service.GetIsuConditionRequest{
					StartTime:        nil,
					CursorEndTime:    uint64(CursorEndTime),
					CursorJIAIsuUUID: "",
					ConditionLevel:   "info,warning,critical",
					Limit:            nil,
				}
				conditionsTmp, res, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
				if err != nil {
					scenarioSuccess = false
					step.AddError(err)
					break
				}
				//検証
				//ここは、古いデータのはずなのでconditionのchanからの再取得は要らない
				err = verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request,
					conditionsTmp, mustExistUntil,
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
			virtualToday := (dataExistTimestamp / (24 * 60 * 60)) * (24 * 60 * 60)
			_, graphToday, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualToday),
				func(res *http.Response, graph []*service.GraphResponse) []error {
					//検証前にデータ取得
					user.GetConditionFromChan(ctx)
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
			_, _, errs = browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, uint64(virtualToday-60*60),
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
				request := service.GetIsuConditionRequest{
					StartTime:        &startTime,
					CursorEndTime:    uint64(errorEndAtUnix),
					CursorJIAIsuUUID: "",
					ConditionLevel:   "warning,critical",
					Limit:            nil,
				}
				_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
					request,
					func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
						//検証
						//ここは、古いデータのはずなのでconditionのchanからの再取得は要らない
						//TODO: starttimeの検証
						err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request,
							conditions, mustExistUntil,
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
