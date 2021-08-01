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
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

func (s *Scenario) Load(parent context.Context, step *isucandar.BenchmarkStep) error {
	defer s.jiaCancel()
	step.Result().Score.Reset()
	if s.NoLoad {
		return nil
	}
	ctx, cancel := context.WithTimeout(parent, s.LoadTimeout)
	defer cancel()

	logger.ContestantLogger.Printf("===> LOAD")
	logger.AdminLogger.Printf("LOAD INFO\n  Language: %s\n  Campaign: None\n", s.Language)
	defer logger.AdminLogger.Println("<=== LOAD END")

	/*
		TODO: 実際の負荷走行シナリオ
	*/

	//通常ユーザー
	s.AddNormalUser(ctx, step, 100)
	//マニアユーザー
	//s.AddManiacUser(ctx, step, 2)
	//企業ユーザー
	s.AddCompanyUser(ctx, step, 5)

	//ユーザーを増やす
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		s.userAdder(ctx, step)
	}()

	<-ctx.Done()
	s.jiaCancel()
	logger.AdminLogger.Println("LOAD WAIT")
	s.loadWaitGroup.Wait()

	return nil
}

// UserLoop を増やすかどうか判定し、増やすなり減らす
func (s *Scenario) userAdder(ctx context.Context, step *isucandar.BenchmarkStep) {
	defer logger.AdminLogger.Println("--- userAdder END")
	//TODO: パラメーター調整
	for {
		select {
		case <-time.After(5000 * time.Millisecond):
		case <-ctx.Done():
			return
		}

		errCount := step.Result().Errors.Count()
		timeoutCount, ok := errCount["timeout"]
		if !ok || timeoutCount == 0 {
			logger.ContestantLogger.Println("現レベルの負荷へ応答ができているため、負荷レベルを上昇させます")
			s.AddNormalUser(ctx, step, 20)
			s.AddCompanyUser(ctx, step, 1)
		} else if ok && timeoutCount > 0 {
			logger.ContestantLogger.Println("エラーが発生したため、負荷レベルは上昇しません")
			return
		}
	}
}

func (s *Scenario) loadNormalUser(ctx context.Context, step *isucandar.BenchmarkStep) {

	userTimer, userTimerCancel := context.WithDeadline(ctx, s.realTimeLoadFinishedAt.Add(-agent.DefaultRequestTimeout))
	defer userTimerCancel()
	select {
	case <-ctx.Done():
		return
	case <-userTimer.Done():
		return
	default:
	}
	// logger.AdminLogger.Println("Normal User start")
	// defer logger.AdminLogger.Println("Normal User END")

	user := s.initNormalUser(ctx, step)
	if user == nil {
		return
	}

	randEngine := rand.New(rand.NewSource(rand.Int63()))
	nextTargetIsuIndex := 0
	scenarioLoopStopper := time.After(1 * time.Millisecond) //ループ頻度調整
	for {
		<-scenarioLoopStopper
		scenarioLoopStopper = time.After(50 * time.Millisecond) //TODO: 頻度調整
		select {
		case <-ctx.Done():
			return
		case <-userTimer.Done(): //TODO: GETリクエスト系も早めに終わるかは要検討
			return
		default:
		}

		//posterからconditionの取得
		user.GetConditionFromChan(ctx)
		select {
		case <-ctx.Done():
			return
		default:
		}

		//conditionを見るISUを選択
		//TODO: 乱数にする
		nextTargetIsuIndex += 1
		nextTargetIsuIndex %= len(user.IsuListOrderByCreatedAt)
		targetIsu := user.IsuListOrderByCreatedAt[nextTargetIsuIndex]

		//GET /
		dataExistTimestamp := GetConditionDataExistTimestamp(s, user)
		_, errs := browserGetHomeAction(ctx, user.Agent, dataExistTimestamp, true,
			func(res *http.Response, isuList []*service.Isu) []error {
				expected := user.IsuListOrderByCreatedAt
				if homeIsuLimit < len(expected) { //limit
					expected = expected[len(expected)-homeIsuLimit:]
				}
				return verifyIsuOrderByCreatedAt(res, expected, isuList)
			},
		)
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		if len(errs) > 0 {
			continue
		}

		//GET /isu/{jia_isu_uuid}
		_, errs = browserGetIsuDetailAction(ctx, user.Agent, targetIsu.JIAIsuUUID, true)
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		if len(errs) > 0 {
			continue
		}

		// TODO: 確定的に
		route := randEngine.Intn(3)
		// 1 / 2
		if route < 1 {
			s.requestNewConditionScenario(ctx, step, user, targetIsu)
		} else if route < 2 {
			s.requestLastBadConditionScenario(ctx, step, user, targetIsu)
		} else {
			s.requestGraphScenario(ctx, step, user, targetIsu, randEngine)
		}
	}
}

//ユーザーとISUの作成
func (s *Scenario) initNormalUser(ctx context.Context, step *isucandar.BenchmarkStep) *model.User {
	//ユーザー作成
	userAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	user := s.NewUser(ctx, step, userAgent, model.UserTypeNormal)
	if user == nil {
		//logger.AdminLogger.Println("Normal User fail: NewUser")
		return nil //致命的でないエラー
	}
	func() {
		s.normalUsersMtx.Lock()
		defer s.normalUsersMtx.Unlock()
		s.normalUsers = append(s.normalUsers, user)
	}()

	//椅子作成
	const isuCountMax = 4 //ルートページに表示する最大数
	isuCount := rand.Intn(isuCountMax) + 1
	for i := 0; i < isuCount; i++ {
		isu := s.NewIsu(ctx, step, user, true, nil)
		// TODO: retry
		if isu == nil {
			return nil
		}
	}
	step.AddScore(ScoreNormalUserInitialize)
	return user
}

// あるISUの新しいconditionを見に行くシナリオ。
func (s *Scenario) requestNewConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu) {
	// 最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	request := service.GetIndividualIsuConditionRequest{
		StartTime:        nil,
		CursorEndTime:    nowVirtualTime.Unix(),
		ConditionLevel:   "info,warning,critical",
		Limit:            nil,
	}
	conditions, errs := s.getIsuConditionUntilAlreadyRead(ctx, user, targetIsu, request)
	if len(errs) != 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		return
	}
	if len(conditions) == 0  {
		return
	}

	// GETに成功したのでその分を加点
	for _, cond := range conditions {
		// TODO: 点数調整考える。ここ読むたびじゃなくて、何件読んだにするとか
		switch cond.ConditionLevel {
		case "info":
			step.AddScore(ScoreReadInfoCondition)
		case "warning":
			step.AddScore(ScoreReadWarningCondition)
		case "critical":
			step.AddScore(ScoreReadCriticalCondition)
		default:
			// validate でここに入らないことは保証されている
		}
	}

	// LastReadConditionTimestamp を更新
	// condition の順番保障はされてる
	targetIsu.LastReadConditionTimestamp = conditions[0].Timestamp

	// このシナリオでは修理しない
}

// あるISUの、悪い最新のconditionを見に行くシナリオ。
func (s *Scenario) requestLastBadConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu) {
	// ConditionLevel最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	request := service.GetIndividualIsuConditionRequest{
		StartTime:        nil,
		CursorEndTime:    nowVirtualTime.Unix(),
		ConditionLevel:   "warning,critical",
		Limit:            nil,
	}
	// GET condition/{jia_isu_uuid} を取得してバリデーション
	_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
			// poster で送ったものの同期
			user.GetConditionFromChan(ctx)

			// TODO: validation は引数に渡さず関数の結果からやる
			//conditionの検証
			err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditions)
			if err != nil {
				return []error{err}
			}
			return []error{}
		},
	)
	// TODO: retry
	if len(errs) > 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
			return
		}
	}
	if len(conditions) == 0 {
		return
	}

	// こっちでは加点しない

	// 新しい condition を確認して、椅子状態を改善
	solveCondition, targetTimestamp := findBadIsuState(conditions)

	// すでに改善してるなら修理とかはしない
	if targetTimestamp <= targetIsu.LastReadBadConditionTimestamp {
		return
	}
	if solveCondition != model.IsuStateChangeNone {
		// 状態改善
		// バッファがあるのでブロック率は低い読みで直列に投げる
		select {
		case <-ctx.Done():
			return
		case targetIsu.StreamsForScenario.StateChan <- solveCondition:
		}
	}
	
	// LastReadBadConditionTimestamp を更新
	// condition の順番保障はされてる
	targetIsu.LastReadBadConditionTimestamp = conditions[0].Timestamp
}

//GET /isu/condition/{jia_isu_uuid} を一度見たconditionが出るまでページングする === 全てが新しいなら次のページに行く。補足: LastReadTimestamp は外で更新
func (s *Scenario) getIsuConditionUntilAlreadyRead(
	ctx context.Context,
	user *model.User,
	targetIsu *model.Isu,
	request service.GetIndividualIsuConditionRequest,
) ([]*service.GetIsuConditionResponse, []error) {
	// 今回のこの関数で取得した condition の配列
	conditions := []*service.GetIsuConditionResponse{}

	// limit を指定しているならそれに合わせて、指定してないならデフォルトの値を使う
	limit := conditionLimit
	if request.Limit != nil {
		limit = *request.Limit
	}

	// GET condition/{jia_isu_uuid} を取得してバリデーション
	_, firstPageConditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
			// poster で送ったものの同期
			user.GetConditionFromChan(ctx)

			// TODO: validation は引数に渡さず関数の結果からやる
			//conditionの検証
			err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditions)
			if err != nil {
				return []error{err}
			}
			return []error{}
		},
	)
	// TODO: retry
	if len(errs) > 0 {
		return nil, errs
	}
	for i, cond := range firstPageConditions {
		// 新しいやつだけなら append
		if isNewData(targetIsu, cond) {
			conditions = append(conditions, cond)
		} else {
			logger.AdminLogger.Printf("is not new data in firstPageConditions. index: %v", i)
			// timestamp順なのは vaidation で保証しているので読んだやつが出てきたタイミングで return
			return conditions, nil
		}
	}
	// 最初のページが最後のページならやめる
	if len(firstPageConditions) != limit {
		return conditions, nil
	}

	// 続きがあり、なおかつ今取得した condition が全て新しい時はスクロールする
	for {
		request = service.GetIndividualIsuConditionRequest{
			StartTime:        request.StartTime,
			CursorEndTime:    conditions[len(conditions)-1].Timestamp,
			ConditionLevel:   request.ConditionLevel,
			Limit:            request.Limit,
		}
		tmpConditions, _, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		// poster で送ったものの同期
		user.GetConditionFromChan(ctx)
		// TODO: retry
		if err != nil {
			return nil, []error{err}
		}
		// TODO: validation

		for _, cond := range tmpConditions {
			// 新しいやつだけなら append
			if isNewData(targetIsu, cond) {
				conditions = append(conditions, cond)
			} else {
				// timestamp順なのは validation で保証しているので読んだやつが出てきたタイミングで return
				return conditions, nil
			}
		}

		// 最後のページまで見ちゃってるならやめる
		if len(tmpConditions) != limit {
			return conditions, nil
		}
	}
}

func isNewData(isu *model.Isu, condition *service.GetIsuConditionResponse) bool {
	return condition.Timestamp > isu.LastReadConditionTimestamp
}

// あるISUのグラフを見に行くシナリオ
func (s *Scenario) requestGraphScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu, randEngine *rand.Rand) {
	// 最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	// 割り算で切り捨てを発生させている(day単位にしている)
	virtualToday := trancateTimestampToDate(nowVirtualTime.Unix())
	virtualToday -= OneDay

	graphResponses, errs := getIsuGraphUntilLastViewed(ctx, user, targetIsu, virtualToday)
	if errs != nil {
		if len(errs) > 0 {
			for _, err := range errs {
				addErrorWithContext(ctx, step, err)
			}
			return
		}
	}

	// LastCompletedGraphTime を更新
	newLastCompletedGraphTime := getNewLastCompletedGraphTime(graphResponses, virtualToday)
	if targetIsu.LastCompletedGraphTime < newLastCompletedGraphTime {
		targetIsu.LastCompletedGraphTime = newLastCompletedGraphTime
	}

	// AddScoreはconditionのGETまで待つためここでタグを持っておく
	scoreTags := []score.ScoreTag{}

	// scoreの計算
	for behindDay, gr := range graphResponses {
		minTimestampCount := int(^uint(0) >> 1) 
		for _, _ = range gr {			
			// TODO: backend側が graph のレスポンスに根拠timestampを追加したらここで以下のコードを実行する
			// if len(g.timestamps) < minTimestampCount {
			// 	minTimestampCount = len(g.timestamps)
			// }
			// ↓こっちの方は消す
			minTimestampCount = 0

		}
		// 「今日のグラフじゃない」&「完成しているグラフ」なら加点
		if behindDay != 0 && targetIsu.LastCompletedGraphTime <= virtualToday - (int64(behindDay) * OneDay) {
			// AddScoreはconditionのGETまで待つためここでタグを入れておく
			scoreTags = append(scoreTags, getGraphScoreTag(minTimestampCount))
		}
	}

	// データが入ってるレスポンスから、ランダムで見る condition を選ぶ
	if len(graphResponses) != 0 {
		// ユーザーが今見ているグラフ
		nowViewingGraph := graphResponses[len(graphResponses) - 1]
		// チェックする時間
		checkHour := getCheckHour(nowViewingGraph, randEngine)
		request := service.GetIndividualIsuConditionRequest{
			StartTime: &nowViewingGraph[checkHour].StartAt,
			CursorEndTime: nowViewingGraph[checkHour].EndAt,
			ConditionLevel: "info,warning,critical",
		}
		_, _ , err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		// TODO: retry
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return
		}
		// TODO: validation
	}

	// graph の加点分を計算
	for _, scoreTag := range scoreTags {
		step.AddScore(scoreTag)
	}
}

// unix timeのtimestampをその「日」に切り捨てる
func trancateTimestampToDate(timestamp int64) int64 {
	return (timestamp / OneDay) * OneDay
}

// 新しい LastCompletedGraphTime を得る。
func getNewLastCompletedGraphTime(graphResponses [][]*service.GraphResponse, virtualToday int64) int64 {
	var lastCompletedGraphTime int64 = 0
	for behindDay, gr := range graphResponses {
		for hour, g := range gr {
			// 12時以降のデータがあるならその前日のグラフは完成している
			if hour >= 12 && g.Data != nil {
				completedDay := virtualToday - (OneDay * int64(behindDay))
				if lastCompletedGraphTime < completedDay {
					lastCompletedGraphTime = completedDay
				}
			}
		}
	}
	return lastCompletedGraphTime
}

// データが入ってる graph のレスポンスから、ランダムでユーザがチェックする condition を選ぶ
func getCheckHour(nowViewingGraph []*service.GraphResponse, randEngine *rand.Rand) int {
	dataExistIndexes := []int{}
	for i, g := range nowViewingGraph {
		if g.Data != nil {
			dataExistIndexes = append(dataExistIndexes, i)
		}
	}
	if len(dataExistIndexes) == 0 {
		return 0
	}
	return randEngine.Intn(len(dataExistIndexes))
}

func getGraphScoreTag(minTimestampCount int) score.ScoreTag {
	if minTimestampCount > ScoreGraphTimestampCount.Excellent {
		return ScoreGraphExcellent
	}
	if minTimestampCount > ScoreGraphTimestampCount.Good {
		return ScoreGraphGood
	}
	if minTimestampCount > ScoreGraphTimestampCount.Normal {
		return ScoreGraphNormal
	}
	if minTimestampCount > ScoreGraphTimestampCount.Bad {
		return ScoreGraphBad
	}
	return ScoreGraphWorst
}

// GET /isu/{jia_isu_uuid}/graph を、「一度見たgraphの次のgraph」or「ベンチがisuの作成を投げた仮想時間の日」まで。補足: LastViewedGraphは外で更新
func getIsuGraphUntilLastViewed(
	ctx context.Context,
	user *model.User,
	targetIsu *model.Isu,
	virtualDay int64,
) ([][]*service.GraphResponse, []error) {
	graph := [][]*service.GraphResponse{}

	// これで今日のグラフを取る
	_, todayGraph, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, virtualDay,
		func(res *http.Response, graph []*service.GraphResponse) []error {
			// TODO: validationは下でやるべき
			//検証前にデータ取得
			user.GetConditionFromChan(ctx)
			return []error{} //TODO: 検証
		},
	)
	if len(errs) > 0 {
		// TODO: timeout 時 retry
		return nil, errs
	}

	graph = append(graph, todayGraph)

	// 見たグラフまでを見に行く
	for {
		// 一日前
		virtualDay -= 24 * 60 * 60
		// すでに見たグラフなら終わる
		if virtualDay == targetIsu.LastCompletedGraphTime {
			return graph, nil
		}

		tmpGraph, _, err := getIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, virtualDay)
		if err != nil {
			return nil, []error{err}
		}
		// TODO: timeoutしたときretry

		// TODO: validation

		graph = append(graph, tmpGraph)

		// 作成した時間まで戻ったら終わる
		if targetIsu.PostTime.Unix() > virtualDay {
			return graph, nil
		}
	}
}

//GET /isu/condition/{jia_isu_uuid} をスクロール付きで取り、バリデーションする
func (s *Scenario) getIsuConditionWithScroll(
	ctx context.Context,
	step *isucandar.BenchmarkStep,
	user *model.User,
	targetIsu *model.Isu,
	request service.GetIndividualIsuConditionRequest,
	scrollCount int,
) []*service.GetIsuConditionResponse {
	//GET condition/{jia_isu_uuid} を取得してバリデーション
	_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
			//conditionの検証
			err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditions)
			if err != nil {
				return []error{err}
			}
			return []error{}
		},
	)
	if len(errs) > 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		return nil
	}

	//続きがある場合はスクロール
	limit := conditionLimit
	if request.Limit != nil {
		limit = *request.Limit
	}
	for i := 0; i < scrollCount && len(conditions) == limit*(i+1); i++ {
		var conditionsTmp []*service.GetIsuConditionResponse
		request = service.GetIndividualIsuConditionRequest{
			StartTime:        request.StartTime,
			CursorEndTime:    conditions[len(conditions)-1].Timestamp,
			ConditionLevel:   request.ConditionLevel,
			Limit:            request.Limit,
		}
		conditionsTmp, res, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return nil
		}
		//検証
		err = verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditionsTmp)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return nil
		}

		conditions = append(conditions, conditionsTmp...)
	}

	return conditions
}

func getIsuGraphWithPaging(
	ctx context.Context,
	step *isucandar.BenchmarkStep,
	user *model.User,
	targetIsu *model.Isu,
	virtualDay int64,
	scrollCount int,
) []*service.GraphResponse {
	//graphを見に行くシナリオ
	_, graph, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, virtualDay,
		func(res *http.Response, graph []*service.GraphResponse) []error {
			//検証前にデータ取得
			user.GetConditionFromChan(ctx)
			return []error{} //TODO: 検証
		},
	)
	for _, err := range errs {
		addErrorWithContext(ctx, step, err)
	}
	if len(errs) > 0 {
		return nil
	}

	//前日,... のグラフ
	for i := 0; i < scrollCount; i++ {
		virtualDay -= 24 * 60 * 60
		_, graphTmp, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, virtualDay,
			func(res *http.Response, graph []*service.GraphResponse) []error {
				return []error{} //TODO: 検証
			},
		)
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		if len(errs) > 0 {
			return nil
		}
		graph = append(graph, graphTmp...)
	}

	return graph
}

func findBadIsuState(conditions []*service.GetIsuConditionResponse) (model.IsuStateChange, int64) {
	var virtualTimestamp int64
	solveCondition := model.IsuStateChangeNone
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
					solveCondition |= model.IsuStateChangeClear
				} else if keyValue[0] == "is_overweight" { // これだけ解消される可能性がある
					solveCondition |= model.IsuStateChangeDetectOverweight
				} else if keyValue[0] == "is_broken" {
					solveCondition |= model.IsuStateChangeRepair
				}
			}
		}
		// TODO: これ == 0 で大丈夫？一度 virtualTimestamp に値を入れた時点で break したほうが良さそう(is_overweight も解消されないようにするなら braek させる)
		if bad && virtualTimestamp == 0 {
			virtualTimestamp = c.Timestamp
		}
	}

	return solveCondition, virtualTimestamp
}

func (s *Scenario) loadCompanyUser(ctx context.Context, step *isucandar.BenchmarkStep) {

	userTimer, userTimerCancel := context.WithDeadline(ctx, s.realTimeLoadFinishedAt.Add(-agent.DefaultRequestTimeout))
	defer userTimerCancel()
	select {
	case <-ctx.Done():
		return
	case <-userTimer.Done():
		return
	default:
	}
	logger.AdminLogger.Println("Company User start")
	defer logger.AdminLogger.Println("Company User END")

	user, _ := s.initCompanyUser(ctx, step)
	if user == nil {
		return //致命的でないエラー
	}

	//椅子作成
	//const isuCountMax = 1000
	isuCount := rand.Intn(10) + 500
	for i := 0; i < isuCount; i++ {
		isu := s.NewIsu(ctx, step, user, true, nil)
		// TODO: retry
		if isu == nil {
			return
		}
	}

	step.AddScore(ScoreCompanyUserInitialize)

	scenarioDoneCount := 0
	scenarioSuccess := false
	lastSolvedTime := make(map[string]time.Time)
	for _, isu := range user.IsuListOrderByCreatedAt {
		lastSolvedTime[isu.JIAIsuUUID] = s.virtualTimeStart
	}
	scenarioLoopStopper := time.After(1 * time.Millisecond)
	for {
		<-scenarioLoopStopper
		scenarioLoopStopper = time.After(50 * time.Millisecond) //TODO: 頻度調整

		select {
		case <-ctx.Done():
			return
		case <-userTimer.Done(): //TODO: GET系リクエストも早めに止めるかは要検討
			return
		default:
		}

		if scenarioSuccess {
			scenarioDoneCount++
			step.AddScore(ScoreCompanyUserLoop) //TODO: 得点条件の修正

			//シナリオに成功している場合は椅子追加
			//TODO: 係数調整
			// for isuCount < (scenarioDoneCount/30)*50 && isuCount < isuCountMax {
			// 	isu := s.NewIsu(ctx, step, user, true)
			// 	if isu == nil {
			// 		logger.AdminLogger.Println("Company User fail: NewIsu")
			// 	} else {
			// 		isuCount++
			// 	}
			// 	//logger.AdminLogger.Printf("Company User Isu: %d\n", isuCount)
			// }
		}
		scenarioSuccess = true

		//posterからconditionの取得
		user.GetConditionFromChan(ctx)
		select {
		case <-ctx.Done():
			return
		default:
		}

		//GET /
		//TODO: ベンチはPUT isu/iconが来ないとして、304を常に許すようにします。
		dataExistTimestamp := GetConditionDataExistTimestamp(s, user)
		_, errs := browserGetHomeAction(ctx, user.Agent, dataExistTimestamp, true,
			func(res *http.Response, isuList []*service.Isu) []error {
				expected := user.IsuListOrderByCreatedAt
				if homeIsuLimit < len(expected) { //limit
					expected = expected[len(expected)-homeIsuLimit:]
				}
				return verifyIsuOrderByCreatedAt(res, expected, isuList)
			},
		)
		for _, err := range errs {
			scenarioSuccess = false
			addErrorWithContext(ctx, step, err)
		}
		if !scenarioSuccess {
			continue
		}

		scenarioSuccess = s.checkCompanyConditionScenario(ctx, step, user, lastSolvedTime)
	}
}

func (s *Scenario) initCompanyUser(ctx context.Context, step *isucandar.BenchmarkStep) (*model.User, []*agent.Agent) {
	//ユーザー作成
	userAgents := make([]*agent.Agent, 10)
	for i := range userAgents {
		var err error
		userAgents[i], err = s.NewAgent()
		if err != nil {
			logger.AdminLogger.Panicln(err)
		}
	}
	user := s.NewUser(ctx, step, userAgents[0], model.UserTypeCompany)
	if user == nil {
		logger.AdminLogger.Println("Company User fail: NewUser")
		return nil, nil
	}
	func() {
		s.companyUsersMtx.Lock()
		defer s.companyUsersMtx.Unlock()
		s.companyUsers = append(s.companyUsers, user)
	}()

	//椅子作成
	//const isuCountMax = 1000
	isuCount := rand.Intn(10) + 500
	for i := 0; i < isuCount; i++ {
		isu := s.NewIsu(ctx, step, user, true, nil)
		// TODO: retry
		if isu == nil {
			return nil, nil
		}
	}

	step.AddScore(ScoreCompanyUserInitialize)
	return user, userAgents
}

func (s *Scenario) checkCompanyConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, lastSolvedTime map[string]time.Time) bool {
	//定期的にconditionを見に行くシナリオ
	scenarioSuccess := true
	dataExistTimestamp := GetConditionDataExistTimestamp(s, user)
	request := service.GetIsuConditionRequest{
		StartTime:        nil,
		CursorEndTime:    dataExistTimestamp,
		ConditionLevel:   "warning,critical",
		Limit:            nil,
	}
	conditions, errs := browserGetConditionsAction(ctx, user.Agent,
		request,
		func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
			// どうせ company は消すからコメントアウト
			//conditionの検証
			// err := verifyIsuConditions(res, user, "", &request, conditions)
			// if err != nil {
			// 	return []error{err}
			// }
			return []error{}
		},
	)
	for _, err := range errs {
		scenarioSuccess = false
		step.AddError(err)
	}
	if len(errs) > 0 || len(conditions) == 0 {
		return scenarioSuccess
	}

	//スクロール
	for i := 0; i < 2 && len(conditions) == 20*(i+1); i++ {
		var conditionsTmp []*service.GetIsuConditionResponse
		request = service.GetIsuConditionRequest{
			StartTime:        nil,
			CursorEndTime:    conditions[len(conditions)-1].Timestamp,
			CursorJIAIsuUUID: conditions[len(conditions)-1].JIAIsuUUID,
			ConditionLevel:   "warning,critical",
			Limit:            nil,
		}
		conditionsTmp, _, err := getConditionAction(ctx, user.Agent, request)
		if err != nil {
			scenarioSuccess = false
			addErrorWithContext(ctx, step, err)
			break
		}
		//検証
		// どうせ消すからコメントアウト
		// err = verifyIsuConditions(res, user, "", &request, conditionsTmp)
		// if err != nil {
		// 	scenarioSuccess = false
		// 	addErrorWithContext(ctx, step, err)
		// 	break
		// }

		conditions = append(conditions, conditionsTmp...)
	}

	//conditionを確認して、椅子状態を改善
	solvedConditions, findTimestamps := findBadIsuStateWithID(conditions)
	for targetIsuID, timestamp := range findTimestamps {
		if solvedConditions[targetIsuID] != model.IsuStateChangeNone && lastSolvedTime[targetIsuID].Before(time.Unix(timestamp, 0)) {
			//graphを見る
			virtualDay := (timestamp / (24 * 60 * 60)) * (24 * 60 * 60)
			_, _, errs := browserGetIsuGraphAction(ctx, user.Agent, targetIsuID, virtualDay,
				func(res *http.Response, graph []*service.GraphResponse) []error {
					return []error{} //TODO: 検証
				},
			)
			for _, err := range errs {
				scenarioSuccess = false
				step.AddError(err)
			}

			//状態改善
			lastSolvedTime[targetIsuID] = time.Unix(timestamp, 0)
			select {
			case <-ctx.Done():
				return false
			case user.IsuListByID[targetIsuID].StreamsForScenario.StateChan <- solvedConditions[targetIsuID]: //バッファがあるのでブロック率は低い読みで直列に投げる
			}
		}
	}

	return scenarioSuccess
}

func findBadIsuStateWithID(conditions []*service.GetIsuConditionResponse) (map[string]model.IsuStateChange, map[string]int64) {
	virtualTimestamp := make(map[string]int64)
	solveCondition := make(map[string]model.IsuStateChange)
	for _, c := range conditions {
		//MEMO: 重かったらフォーマットが想定通りの前提で最適化する
		bad := false
		if _, ok := solveCondition[c.JIAIsuUUID]; !ok {
			solveCondition[c.JIAIsuUUID] = model.IsuStateChangeNone
		}
		for _, cond := range strings.Split(c.Condition, ",") {
			keyValue := strings.Split(cond, "=")
			if len(keyValue) != 2 {
				continue //形式に従っていないものは無視
			}
			if keyValue[1] != "false" {
				bad = true
				if keyValue[0] == "is_dirty" {
					solveCondition[c.JIAIsuUUID] |= model.IsuStateChangeClear
				} else if keyValue[0] == "is_overweight" {
					solveCondition[c.JIAIsuUUID] |= model.IsuStateChangeDetectOverweight
				} else if keyValue[0] == "is_broken" {
					solveCondition[c.JIAIsuUUID] |= model.IsuStateChangeRepair
				}
			}
		}
		if _, ok := virtualTimestamp[c.JIAIsuUUID]; bad && !ok {
			virtualTimestamp[c.JIAIsuUUID] = c.Timestamp
		}
	}

	return solveCondition, virtualTimestamp
}
