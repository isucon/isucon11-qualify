package scenario

// load.go
// シナリオの内、loadフェーズの処理

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

var (
	// ユーザーが持つ ISU の数を確定させたいので、そのための乱数生成器。ソースは適当に決めた
	isuCountRandEngine      = rand.New(rand.NewSource(-8679036))
	isuCountRandEngineMutex sync.Mutex

	// 全ユーザーがよんだconditionの端数の合計。Goroutine終了時に加算する
	readInfoConditionFraction     int32 = 0
	readWarnConditionFraction     int32 = 0
	readCriticalConditionFraction int32 = 0

	// Viewer が増やす更新された
	viewUpdatedTrendCounter int32 = 0

	// user loop の数
	userLoopCount int32 = 0
)

type ReadConditionCount struct {
	Info     int32
	Warn     int32
	Critical int32
}

func (s *Scenario) Load(parent context.Context, step *isucandar.BenchmarkStep) error {
	step.Result().Score.Reset()
	if s.NoLoad {
		return nil
	}
	ctx, cancel := context.WithTimeout(parent, s.LoadTimeout)
	defer cancel()

	// // 初期データをロード
	// logger.AdminLogger.Println("start: load initial data")
	// s.InitializeData(ctx)
	// logger.AdminLogger.Println("finish: load initial data")

	logger.ContestantLogger.Printf("===> LOAD")
	logger.AdminLogger.Printf("LOAD INFO\n  Language: %s\n  Campaign: None\n", s.Language)
	defer logger.AdminLogger.Println("<=== LOAD END")

	// 実際の負荷走行シナリオ

	//通常ユーザー
	s.AddNormalUser(ctx, step, 1)
	s.AddIsuconUser(ctx, step)

	//非ログインユーザーを増やす
	//s.AddViewer(ctx, step, 2)
	//ユーザーを増やす
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		s.userAdder(ctx, step)
	}()

	<-ctx.Done()
	s.JiaCancel()
	logger.AdminLogger.Println("LOAD WAIT")
	s.loadWaitGroup.Wait()

	// 余りの加点
	addConditionScoreTag(step, &ReadConditionCount{
		Info:     atomic.LoadInt32(&readInfoConditionFraction),
		Warn:     atomic.LoadInt32(&readWarnConditionFraction),
		Critical: atomic.LoadInt32(&readCriticalConditionFraction),
	})

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
		if !ok {
			timeoutCount = 0
		}

		if int32(timeoutCount) > TimeoutLimitPerUser*atomic.LoadInt32(&userLoopCount) {
			logger.ContestantLogger.Println("タイムアウト数が上限に達したため、以降負荷レベルは上昇しません")
			break
		}

		addStep := AddUserStep * atomic.LoadInt32(&userLoopCount)
		addCount := atomic.LoadInt32(&viewUpdatedTrendCounter) / addStep
		if addCount > 0 {
			logger.ContestantLogger.Printf("現レベルの負荷へ応答ができているため、負荷レベルを%d上昇させます", addCount)
			s.AddNormalUser(ctx, step, AddUserCount*int(addCount))
			atomic.AddInt32(&viewUpdatedTrendCounter, -addStep*addCount)
		} else {
			logger.ContestantLogger.Println("負荷レベルは上昇しませんでした")
		}
	}
}

func (s *Scenario) loadNormalUser(ctx context.Context, step *isucandar.BenchmarkStep, isIsuconUser bool) {
	atomic.AddInt32(&userLoopCount, 1)

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

	user := s.initNormalUser(ctx, step, isIsuconUser)
	if user == nil {
		return
	}
	defer user.CloseAllIsuStateChan()

	step.AddScore(ScoreNormalUserInitialize)

	readConditionCount := ReadConditionCount{Info: 0, Warn: 0, Critical: 0}
	defer func() {
		atomic.AddInt32(&readInfoConditionFraction, readConditionCount.Info)
		atomic.AddInt32(&readWarnConditionFraction, readConditionCount.Warn)
		atomic.AddInt32(&readCriticalConditionFraction, readConditionCount.Critical)
	}()

	randEngine := rand.New(rand.NewSource(rand.Int63()))
	nextTargetIsuIndex := 0
	nextScenarioIndex := 0
	scenarioLoopStopper := time.After(1 * time.Millisecond) //ループ頻度調整
	loopCount := 0
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

		select {
		case <-ctx.Done():
			return
		default:
		}

		// 一つのISUに対するシナリオが終わっているとき
		if nextScenarioIndex > 2 {
			//conditionを見るISUを選択
			//TODO: 乱数にする
			nextTargetIsuIndex += 1
			nextTargetIsuIndex %= len(user.IsuListOrderByCreatedAt)
			nextScenarioIndex = 0

			loopCount++

			if loopCount%ViewerAddLoopStep == 0 {
				s.viewerMtx.Lock()
				// 「1 set のシナリオが ViewerAddLoopStep 回終わった」＆「 viewer が ユーザー数×ViewerLimitPerUser 以下」なら Viewer を増やす
				if len(s.viewers) < int(atomic.LoadInt32(&userLoopCount))*ViewerLimitPerUser {
					s.AddViewer(ctx, step, 1)
				}
				s.viewerMtx.Unlock()
			}
		}
		targetIsu := user.IsuListOrderByCreatedAt[nextTargetIsuIndex]

		//GET /
		var newConditionUUIDs []string
		_, errs := browserGetHomeAction(ctx, user.Agent,
			func(res *http.Response, isuList []*service.Isu) []error {
				expected := user.IsuListOrderByCreatedAt

				var errs []error
				newConditionUUIDs, errs = verifyIsuList(res, expected, isuList)
				return errs
			},
		)
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		if len(errs) > 0 {
			continue
		}
		//更新されているかどうか確認
		if nextScenarioIndex == 0 {
			found := false
			for _, updated := range newConditionUUIDs {
				if updated == targetIsu.JIAIsuUUID {
					found = true
					break
				}
			}
			if !found { //更新されていないので次のISUを見に行く
				nextScenarioIndex = 3
				continue
			}
		}

		//GET /isu/{jia_isu_uuid}
		_, errs = browserGetIsuDetailAction(ctx, user.Agent, targetIsu.JIAIsuUUID)
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		if len(errs) > 0 {
			continue
		}

		if nextScenarioIndex == 0 {
			s.requestNewConditionScenario(ctx, step, user, targetIsu, &readConditionCount)
		} else if nextScenarioIndex == 1 {
			s.requestLastBadConditionScenario(ctx, step, user, targetIsu)
		} else {
			s.requestGraphScenario(ctx, step, user, targetIsu, randEngine)
		}

		// たまに signoutScenario に入る
		if randEngine.Intn(100) < SignoutPercentage {
			signoutScenario(ctx, step, user)
		}

		// 次のシナリオに
		nextScenarioIndex += 1
	}
}

func (s *Scenario) loadViewer(ctx context.Context, step *isucandar.BenchmarkStep) {

	viewerTimer, viewerTimerCancel := context.WithDeadline(ctx, s.realTimeLoadFinishedAt.Add(-agent.DefaultRequestTimeout))
	defer viewerTimerCancel()
	select {
	case <-ctx.Done():
		return
	case <-viewerTimer.Done():
		return
	default:
	}

	viewer := s.initViewer(ctx)
	scenarioLoopStopper := time.After(1 * time.Millisecond) //ループ頻度調整
	for {
		<-scenarioLoopStopper
		scenarioLoopStopper = time.After(10 * time.Millisecond)

		select {
		case <-ctx.Done():
			return
		case <-viewerTimer.Done(): //TODO: GETリクエスト系も早めに終わるかは要検討
			return
		default:
		}

		// viewer が ViewerDropCount 以上エラーに遭遇していたらループから脱落
		if viewer.ErrorCount >= ViewerDropCount {
			return
		}

		requestTime := time.Now()
		trend, res, err := getTrendAction(ctx, viewer.Agent)
		if err != nil {
			viewer.ErrorCount += 1
			addErrorWithContext(ctx, step, err)
			continue
		}
		updatedCount, err := s.verifyTrend(ctx, res, viewer, trend, requestTime)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			viewer.ErrorCount += 1
			continue
		}
		atomic.AddInt32(&viewUpdatedTrendCounter, int32(updatedCount))
	}
}

// ユーザーとISUの作成
func (s *Scenario) initNormalUser(ctx context.Context, step *isucandar.BenchmarkStep, isIsuconUser bool) *model.User {
	//ユーザー作成
	userAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	user := s.NewUser(ctx, step, userAgent, model.UserTypeNormal, isIsuconUser)
	if user == nil {
		//logger.AdminLogger.Println("Normal User fail: NewUser")
		return nil //致命的でないエラー
	}
	func() {
		s.normalUsersMtx.Lock()
		defer s.normalUsersMtx.Unlock()
		s.normalUsers = append(s.normalUsers, user)
	}()

	// s.NewUser() 内で POST /api/auth をしているためトップページに飛ぶ
	_, errs := browserGetHomeAction(ctx, user.Agent,
		func(res *http.Response, isuList []*service.Isu) []error {
			expected := user.IsuListOrderByCreatedAt

			var errs []error
			_, errs = verifyIsuList(res, expected, isuList)
			return errs
		},
	)
	for _, err := range errs {
		addErrorWithContext(ctx, step, err)
		// 致命的なエラーではないため return しない
	}

	//椅子作成
	// TODO: 実際に解いてみてこの isu 数の上限がいい感じに働いているか検証する
	isuCountRandEngineMutex.Lock()
	isuCount := isuCountRandEngine.Intn(IsuCountMax) + 1
	isuCountRandEngineMutex.Unlock()

	for i := 0; i < isuCount; i++ {
		image, err := random.Image()
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		isu := s.NewIsu(ctx, step, user, true, image, true)
		if isu == nil {
			user.CloseAllIsuStateChan()
			return nil
		}
	}

	return user
}

//ユーザーとISUの作成
func (s *Scenario) initViewer(ctx context.Context) model.Viewer {
	//ユーザー作成
	viewerAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	viewer := model.NewViewer(viewerAgent)
	func() {
		s.viewerMtx.Lock()
		defer s.viewerMtx.Unlock()
		s.viewers = append(s.viewers, &viewer)
	}()

	return viewer
}

// あるISUの新しいconditionを見に行くシナリオ。
func (s *Scenario) requestNewConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu, readConditionCount *ReadConditionCount) {
	// 最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	request := service.GetIsuConditionRequest{
		StartTime:      nil,
		EndTime:        nowVirtualTime.Unix(),
		ConditionLevel: "info,warning,critical",
	}
	conditions, newLastReadConditionTimestamp, errs := s.getIsuConditionUntilAlreadyRead(ctx, user, targetIsu, request, step, readConditionCount)
	if len(errs) != 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		return
	}

	// GETに成功したのでその分を加点
	readCondition(conditions, step, readConditionCount)

	// LastReadConditionTimestamp を更新
	if targetIsu.LastReadConditionTimestamp < newLastReadConditionTimestamp {
		targetIsu.LastReadConditionTimestamp = newLastReadConditionTimestamp
	}

	// このシナリオでは修理しない
}

// あるISUの、悪い最新のconditionを見に行くシナリオ。
func (s *Scenario) requestLastBadConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu) {
	// ConditionLevel最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	request := service.GetIsuConditionRequest{
		StartTime:      nil,
		EndTime:        nowVirtualTime.Unix(),
		ConditionLevel: "warning,critical",
	}
	// GET condition/{jia_isu_uuid} を取得してバリデーション
	_, conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
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
	request service.GetIsuConditionRequest,
	step *isucandar.BenchmarkStep,
	readConditionCount *ReadConditionCount,
) ([]*service.GetIsuConditionResponse, int64, []error) {
	// 更新用のLastReadConditionTimestamp
	var newLastReadConditionTimestamp int64 = 0

	// 今回のこの関数で取得した condition の配列
	conditions := []*service.GetIsuConditionResponse{}

	// GET condition/{jia_isu_uuid} を取得してバリデーション
	_, firstPageConditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions []*service.GetIsuConditionResponse) []error {
			err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditions)
			if err != nil {
				return []error{err}
			}
			return []error{}
		},
	)
	if len(errs) > 0 {
		return nil, newLastReadConditionTimestamp, errs
	}
	if len(firstPageConditions) > 0 {
		newLastReadConditionTimestamp = firstPageConditions[0].Timestamp
	}
	for _, cond := range firstPageConditions {
		// 新しいやつだけなら append
		if isNewData(targetIsu, cond) {
			conditions = append(conditions, cond)
		} else {
			// timestamp順なのは validation で保証しているので読んだやつが出てきたタイミングで return
			return conditions, newLastReadConditionTimestamp, nil
		}
	}
	// 最初のページが最後のページならやめる
	if len(firstPageConditions) < conditionLimit {
		return conditions, newLastReadConditionTimestamp, nil
	}

	pagingCount := 1
	// 続きがあり、なおかつ今取得した condition が全て新しい時はスクロールする
	for {
		request = service.GetIsuConditionRequest{
			StartTime:      request.StartTime,
			EndTime:        conditions[len(conditions)-1].Timestamp,
			ConditionLevel: request.ConditionLevel,
		}

		// ConditionPagingStep ページごとに現状の condition をスコアリング
		pagingCount++
		if pagingCount%ConditionPagingStep == 0 {
			readCondition(conditions, step, readConditionCount)
			conditions = conditions[:0]
		}

		tmpConditions, hres, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			return nil, newLastReadConditionTimestamp, []error{err}
		}
		err = verifyIsuConditions(hres, user, targetIsu.JIAIsuUUID, &request, tmpConditions)
		if err != nil {
			return nil, newLastReadConditionTimestamp, []error{err}
		}

		for _, cond := range tmpConditions {
			// 新しいやつだけなら append
			if isNewData(targetIsu, cond) {
				conditions = append(conditions, cond)
			} else {
				// timestamp順なのは validation で保証しているので読んだやつが出てきたタイミングで return
				return conditions, newLastReadConditionTimestamp, nil
			}
		}

		// 最後のページまで見ちゃってるならやめる
		if len(tmpConditions) != conditionLimit {
			return conditions, newLastReadConditionTimestamp, nil
		}
	}
}

func readCondition(conditions []*service.GetIsuConditionResponse, step *isucandar.BenchmarkStep, readConditionCount *ReadConditionCount) {
	for _, condition := range conditions {
		switch condition.ConditionLevel {
		case "info":
			readConditionCount.Info += 1
		case "warning":
			readConditionCount.Warn += 1
		case "critical":
			readConditionCount.Critical += 1
		default:
			// validate でここに入らないことは保証されている
		}
	}

	addConditionScoreTag(step, readConditionCount)
}

func addConditionScoreTag(step *isucandar.BenchmarkStep, readConditionCount *ReadConditionCount) {
	if readConditionCount.Info-ReadConditionTagStep >= 0 {
		addCount := int(readConditionCount.Info / ReadConditionTagStep)
		for i := 0; i < addCount; i++ {
			step.AddScore(ScoreReadInfoCondition)
			readConditionCount.Info -= ReadConditionTagStep
		}
	}
	if readConditionCount.Warn-ReadConditionTagStep >= 0 {
		addCount := int(readConditionCount.Warn / ReadConditionTagStep)
		for i := 0; i < addCount; i++ {
			step.AddScore(ScoreReadWarningCondition)
			readConditionCount.Warn -= ReadConditionTagStep
		}
	}
	if readConditionCount.Critical-ReadConditionTagStep >= 0 {
		addCount := int(readConditionCount.Critical / ReadConditionTagStep)
		for i := 0; i < addCount; i++ {
			step.AddScore(ScoreReadCriticalCondition)
			readConditionCount.Critical -= ReadConditionTagStep
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
	if len(errs) > 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		return
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
		for _, g := range *gr {
			// 「今日のグラフ」＆「リクエストした時間より先」ならもう minTimestampCount についてカウントしない
			if behindDay == 0 && nowVirtualTime.Unix() < g.EndAt {
				break
			}
			if len(g.ConditionTimestamps) < minTimestampCount {
				minTimestampCount = len(g.ConditionTimestamps)
			}
		}
		// 「今日のグラフじゃない」&「完成しているグラフ」なら加点
		if behindDay != 0 && targetIsu.LastCompletedGraphTime <= virtualToday-(int64(behindDay)*OneDay) {
			// AddScoreはconditionのGETまで待つためここでタグを入れておく
			scoreTags = append(scoreTags, getGraphScoreTag(minTimestampCount))
		}
		// 「今日のグラフ」についても加点
		if behindDay == 0 {
			// AddScoreはconditionのGETまで待つためここでタグを入れておく
			scoreTags = append(scoreTags, getGraphScoreTag(minTimestampCount))
		}
	}

	// データが入ってるレスポンスから、ランダムで見る condition を選ぶ
	if len(graphResponses) != 0 {
		// ユーザーが今見ているグラフ
		nowViewingGraph := graphResponses[len(graphResponses)-1]
		// チェックする時間
		checkHour := getCheckHour(*nowViewingGraph, randEngine)
		request := service.GetIsuConditionRequest{
			StartTime:      &(*nowViewingGraph)[checkHour].StartAt,
			EndTime:        (*nowViewingGraph)[checkHour].EndAt,
			ConditionLevel: "info,warning,critical",
		}
		conditions, hres, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return
		}
		err = verifyIsuConditions(hres, user, targetIsu.JIAIsuUUID, &request, conditions)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return
		}
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
func getNewLastCompletedGraphTime(graphResponses []*service.GraphResponse, virtualToday int64) int64 {
	var lastCompletedGraphTime int64 = 0
	for behindDay, gr := range graphResponses {
		for hour, g := range *gr {
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
func getCheckHour(nowViewingGraph service.GraphResponse, randEngine *rand.Rand) int {
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
) ([]*service.GraphResponse, []error) {
	graph := []*service.GraphResponse{}

	todayRequest := service.GetGraphRequest{Date: virtualDay}
	todayGraph, hres, err := getIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, todayRequest)
	if err != nil {
		return nil, []error{err}
	}

	err = verifyGraph(hres, user, targetIsu.JIAIsuUUID, &todayRequest, todayGraph)
	if err != nil {
		return nil, []error{err}
	}

	graph = append(graph, &todayGraph)

	// 見たグラフまでを見に行く
	for {
		// 一日前
		virtualDay -= 24 * 60 * 60
		// すでに見たグラフなら終わる
		if virtualDay == targetIsu.LastCompletedGraphTime {
			return graph, nil
		}

		request := service.GetGraphRequest{Date: virtualDay}

		tmpGraph, hres, err := getIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			return nil, []error{err}
		}
		err = verifyGraph(hres, user, targetIsu.JIAIsuUUID, &request, tmpGraph)
		if err != nil {
			return nil, []error{err}
		}

		graph = append(graph, &tmpGraph)

		// 作成した時間まで戻ったら終わる
		if targetIsu.PostTime.Unix() > virtualDay {
			return graph, nil
		}
	}
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
		// TODO: これ == 0 で大丈夫？一度 virtualTimestamp に値を入れた時点で break したほうが良さそう(is_overweight も解消されないようにするなら break させる)
		if bad && virtualTimestamp == 0 {
			virtualTimestamp = c.Timestamp
		}
	}

	return solveCondition, virtualTimestamp
}

// 確率で signout して再度ログインするシナリオ。全てのシナリオの最後に確率で発生する
func signoutScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User) {
	_, err := signoutAction(ctx, user.Agent)
	if err != nil {
		addErrorWithContext(ctx, step, err)
		// MEMO: ここで実は signout に成功していました、みたいな状況だと以降のこのユーザーループが死ぬがそれはユーザー責任とする
		return
	}

	// signout したらトップページに飛ぶ(MEMO: 初期状態だと trend おもすぎて backend をころしてしまうかも)
	go func() {
		// 登録済みユーザーは trend に興味はないので verify はせず投げっぱなし
		_, err = getTrendIgnoreAction(ctx, user.Agent)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			// return するとこのあとのログイン必須なシナリオが回らないから return はしない
		}
	}()

	authInfinityRetry(ctx, user.Agent, user.UserID, step)
}

// signoutScenario 以外からは呼ばない(シナリオループの最後である必要がある)
func authInfinityRetry(ctx context.Context, a *agent.Agent, userID string, step *isucandar.BenchmarkStep) {
	for {
		select {
		case <-ctx.Done():
			// 失敗したときも区別せずに return してよい(次シナリオループで終了するため)
			return
		default:
		}
		_, errs := authAction(ctx, a, userID)
		if len(errs) > 0 {
			for _, err := range errs {
				addErrorWithContext(ctx, step, err)
			}
			continue
		}
		return
	}
}

func postIsuInfinityRetry(ctx context.Context, a *agent.Agent, req service.PostIsuRequest, step *isucandar.BenchmarkStep) *http.Response {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		_, res, err := postIsuAction(ctx, a, req)
		if err != nil {
			if res != nil && res.StatusCode == http.StatusConflict {
				return res
			}
			addErrorWithContext(ctx, step, err)
			continue
		}
		return res
	}
}

func getIsuInfinityRetry(ctx context.Context, a *agent.Agent, id string, step *isucandar.BenchmarkStep) (*service.Isu, *http.Response) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}
		isu, res, err := getIsuIdAction(ctx, a, id)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			continue
		}
		return isu, res
	}
}
