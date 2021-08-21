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

	// Viewer の制限
	viewerLimiter chan struct{} = make(chan struct{})

	userAdderIsDropped = make(chan struct{})
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
	// 初期実装だと fail してしまうため下駄をはかせる
	step.AddScore(ScoreStartBenchmark)

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
	s.AddNormalUser(ctx, step, 6)
	s.AddIsuconUser(ctx, step)

	//非ログインユーザーを増やす
	//s.AddViewer(ctx, step, 2)
	//ユーザーを増やす
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() userAdder")
		s.userAdder(ctx, step)
	}()

	//postした件数を記録
	//s.loadWaitGroup.Add(1)
	go func() {
		//defer s.loadWaitGroup.Done()
		//defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() postConditionNumReporter")
		s.postConditionNumReporter(ctx, step)
	}()

	//不正なcondition
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() keepPostingError")
		s.keepPostingError(ctx)
	}()

	//prepare相当のチェック
	s.loadWaitGroup.Add(1)
	go func() {
		defer s.loadWaitGroup.Done()
		defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() loadErrorCheck")
		s.loadErrorCheck(ctx, step)
	}()

	<-ctx.Done()
	s.JiaPosterCancel()
	logger.AdminLogger.Println("LOAD WAIT")

	loadWaitCh := make(chan struct{}, 1)

	go func() {
		s.loadWaitGroup.Wait()
		close(loadWaitCh)
	}()

	select {
	case <-time.After(100 * time.Millisecond):
		// 1秒だけ待ったら抜ける
		logger.AdminLogger.Println("WARNING!!: Force ending loadWaitGroup")
	case <-loadWaitCh:
		// あるいは、正しく loadWaitGroup が done したら抜ける
		logger.AdminLogger.Println("end s.loadWaitGroup.Wait()")
	}

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
	defer func() {
		close(userAdderIsDropped)
		logger.AdminLogger.Println("--- userAdder END")
	}()
	for {
		select {
		case <-time.After(5000 * time.Millisecond):
		case <-ctx.Done():
			return
		}
		userLoopCountLocal := atomic.LoadInt32(&userLoopCount)
		if userLoopCountLocal == 0 {
			continue
		}

		errCount := step.Result().Errors.Count()
		timeoutCount, ok := errCount["timeout"]
		if !ok {
			timeoutCount = 0
		}

		if int32(timeoutCount) > TimeoutLimitPerUser*userLoopCountLocal {
			logger.ContestantLogger.Println("タイムアウト数が多いため、サービスの評判が悪くなりました。以降ユーザーは増加しません")
			break
		}

		addStep := AddUserStep * userLoopCountLocal
		addCount := atomic.LoadInt32(&viewUpdatedTrendCounter) / addStep
		if addCount > 0 {
			logger.ContestantLogger.Printf("サービスの評判が良くなり、ユーザーが%d人増えました", AddUserCount*int(addCount))
			s.AddNormalUser(ctx, step, AddUserCount*int(addCount))
			atomic.AddInt32(&viewUpdatedTrendCounter, -addStep*addCount)
		} else {
			logger.ContestantLogger.Println("ユーザーは増えませんでした")
		}
	}
}

func (s *Scenario) loadNormalUser(ctx context.Context, step *isucandar.BenchmarkStep, isIsuconUser bool) {
	atomic.AddInt32(&userLoopCount, 1)
	go func() {
		// 「1 set のシナリオが ViewerAddLoopStep 回終わった」＆「 viewer が ユーザー数×ViewerLimitPerUser 以下」なら Viewer を増やす
		for i := 0; i < ViewerLimitPerUser; i++ {
			viewerLimiter <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
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
		scenarioLoopStopper = time.After(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 一つのISUに対するシナリオが終わっているとき
		if nextScenarioIndex > 2 {
			//conditionを見るISUを選択
			// できるだけガチャにならないように順番は確定でやる
			nextTargetIsuIndex += 1
			nextTargetIsuIndex %= len(user.IsuListOrderByCreatedAt)
			nextScenarioIndex = 0

			loopCount++

			if loopCount%ViewerAddLoopStep == 0 {
				s.AddViewer(ctx, step, 1)
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
		_, errs = browserGetIsuDetailAction(ctx, user.Agent, targetIsu.JIAIsuUUID, func(res *http.Response, isu *service.Isu) []error {
			errs := []error{}
			err := verifyIsu(res, targetIsu, isu)
			if err != nil {
				errs = append(errs, err)
			}
			// isu.Icon が nil じゃないときはすでにエラーを追加している
			if isu.Icon != nil {
				err = verifyIsuIcon(targetIsu, isu.Icon, isu.IconStatusCode)
				if err != nil {
					errs = append(errs, err)
				}
			}
			return errs
		})
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		if len(errs) > 0 {
			continue
		}

		var isSuccess bool
		if nextScenarioIndex == 0 {
			isSuccess = s.requestNewConditionScenario(ctx, step, user, targetIsu, &readConditionCount)
		} else if nextScenarioIndex == 1 {
			isSuccess = s.requestLastBadConditionScenario(ctx, step, user, targetIsu)
		} else {
			isSuccess = s.requestGraphScenario(ctx, step, user, targetIsu, randEngine)
		}

		// たまに signoutScenario に入る
		if randEngine.Intn(100) < SignoutPercentage {
			signoutScenario(ctx, step, user)
		}

		if isSuccess {
			// 次のシナリオに
			nextScenarioIndex += 1
		}
	}
}

func (s *Scenario) loadViewer(ctx context.Context, step *isucandar.BenchmarkStep) {

	select {
	case <-ctx.Done():
		return
	case <-userAdderIsDropped:
		return
	case <-viewerLimiter:
	}

	viewer := s.initViewer(ctx)
	step.AddScore(ScoreViewerInitialize)
	scenarioLoopStopper := time.After(1 * time.Millisecond) //ループ頻度調整
	for {
		<-scenarioLoopStopper
		scenarioLoopStopper = time.After(100 * time.Millisecond)

		select {
		case <-ctx.Done():
			return
		case <-userAdderIsDropped:
			return //ユーザーが増えなくなったので脱落
		default:
		}

		// viewer が ViewerDropCount 以上エラーに遭遇していたらループから脱落
		if viewer.ErrorCount >= ViewerDropCount {
			step.AddScore(ScoreViewerDropout)
			return
		}

		requestTime := time.Now()
		trend, res, errs := browserGetLandingPageAction(ctx, viewer)
		if len(errs) != 0 {
			viewer.ErrorCount += 1
			for _, err := range errs {
				addErrorWithContext(ctx, step, err)
			}
			continue
		}
		updatedCount, err := s.verifyTrend(ctx, res, viewer, trend, requestTime)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			viewer.ErrorCount += 1
			continue
		}
		atomic.AddInt32(&viewUpdatedTrendCounter, int32(updatedCount))
		step.AddScore(ScoreViewerLoop)
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
		return nil
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
	isuCountRandEngineMutex.Lock()
	isuCount := isuCountRandEngine.Intn(IsuCountMax) + 1
	isuCountRandEngineMutex.Unlock()

	for i := 0; i < isuCount; i++ {
		isu := s.NewIsu(ctx, step, user, true, true)
		if isu == nil {
			user.CloseAllIsuStateChan()
			return nil
		}
		step.AddScore(ScoreIsuInitialize)
	}

	atomic.StoreInt32(&user.PostIsuFinish, 1)

	return user
}

//ユーザーとISUの作成
func (s *Scenario) initViewer(ctx context.Context) *model.Viewer {
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

	return &viewer
}

// あるISUの新しいconditionを見に行くシナリオ。
func (s *Scenario) requestNewConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu, readConditionCount *ReadConditionCount) bool {
	// 最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	request := service.GetIsuConditionRequest{
		StartTime:      nil,
		EndTime:        nowVirtualTime.Unix(),
		ConditionLevel: "info,warning,critical",
	}
	conditions, newLastReadConditionTimestamps, errs := s.getIsuConditionUntilAlreadyRead(ctx, user, targetIsu, request, step, readConditionCount)

	// LastReadConditionTimestamp を更新
	var nextTimestamps [service.ConditionLimit]int64
	indexIsu := 0
	indexNew := 0
	for i := 0; i < service.ConditionLimit; i++ {
		if targetIsu.LastReadConditionTimestamps[indexIsu] == newLastReadConditionTimestamps[indexNew] {
			nextTimestamps[i] = targetIsu.LastReadConditionTimestamps[indexIsu]
			indexIsu++
			indexNew++
		} else if targetIsu.LastReadConditionTimestamps[indexIsu] < newLastReadConditionTimestamps[indexNew] {
			nextTimestamps[i] = newLastReadConditionTimestamps[indexNew]
			indexNew++
		} else {
			nextTimestamps[i] = targetIsu.LastReadConditionTimestamps[indexIsu]
			indexIsu++
		}
	}
	targetIsu.LastReadConditionTimestamps = nextTimestamps

	if len(errs) != 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		return false
	}
	// GETに成功したのでその分を加点
	readCondition(conditions, step, readConditionCount)

	// このシナリオでは修理しない
	return true
}

// あるISUの、悪い最新のconditionを見に行くシナリオ。
func (s *Scenario) requestLastBadConditionScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu) bool {
	// ConditionLevel最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	request := service.GetIsuConditionRequest{
		StartTime:      nil,
		EndTime:        nowVirtualTime.Unix(),
		ConditionLevel: "warning,critical",
	}

	requestTimeUnix := time.Now().Unix()
	// GET condition/{jia_isu_uuid} を取得してバリデーション
	conditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions service.GetIsuConditionResponseArray) []error {
			err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditions, targetIsu.LastReadBadConditionTimestamps, requestTimeUnix)
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
		return false
	}
	if len(conditions) == 0 {
		return true
	}

	// こっちでは加点しない

	// 新しい condition を確認して、椅子状態を改善
	solveCondition, targetTimestamp := findBadIsuState(conditions)

	// すでに改善してるなら修理とかはしない
	if targetTimestamp <= targetIsu.LastReadBadConditionTimestamps[0] {
		return true
	}
	if solveCondition != model.IsuStateChangeNone {
		// 状態改善
		// バッファがあるのでブロック率は低い読みで直列に投げる
		select {
		case <-ctx.Done():
			return false
		case targetIsu.StreamsForScenario.StateChan <- solveCondition:
		}
		step.AddScore(ScoreRepairIsu)
	}

	// LastReadBadConditionTimestamp を更新
	// condition の順番保障はされてる
	var nextTimestamps [service.ConditionLimit]int64
	indexIsu := 0
	indexNew := 0
	for i := 0; i < service.ConditionLimit; i++ {
		if indexNew < len(conditions) && targetIsu.LastReadBadConditionTimestamps[indexIsu] == conditions[indexNew].Timestamp {
			nextTimestamps[i] = targetIsu.LastReadBadConditionTimestamps[indexIsu]
			indexIsu++
			indexNew++
		} else if indexNew < len(conditions) && targetIsu.LastReadBadConditionTimestamps[indexIsu] < conditions[indexNew].Timestamp {
			nextTimestamps[i] = conditions[indexNew].Timestamp
			indexNew++
		} else {
			nextTimestamps[i] = targetIsu.LastReadBadConditionTimestamps[indexIsu]
			indexIsu++
		}
	}
	targetIsu.LastReadBadConditionTimestamps = nextTimestamps

	return true
}

//GET /isu/condition/{jia_isu_uuid} を一度見たconditionが出るまでページングする === 全てが新しいなら次のページに行く。補足: LastReadTimestamp は外で更新
func (s *Scenario) getIsuConditionUntilAlreadyRead(
	ctx context.Context,
	user *model.User,
	targetIsu *model.Isu,
	request service.GetIsuConditionRequest,
	step *isucandar.BenchmarkStep,
	readConditionCount *ReadConditionCount,
) (service.GetIsuConditionResponseArray, [service.ConditionLimit]int64, []error) {
	// 更新用のLastReadConditionTimestamp
	var newLastReadConditionTimestamps [service.ConditionLimit]int64

	// 今回のこの関数で取得した condition の配列
	conditions := service.GetIsuConditionResponseArray{}

	requestTimeUnix := time.Now().Unix()
	// GET condition/{jia_isu_uuid} を取得してバリデーション
	firstPageConditions, errs := browserGetIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID,
		request,
		func(res *http.Response, conditions service.GetIsuConditionResponseArray) []error {
			err := verifyIsuConditions(res, user, targetIsu.JIAIsuUUID, &request, conditions, targetIsu.LastReadConditionTimestamps, requestTimeUnix)
			if err != nil {
				return []error{err}
			}
			return []error{}
		},
	)
	if len(errs) > 0 {
		return nil, newLastReadConditionTimestamps, errs
	}
	for i, c := range firstPageConditions {
		newLastReadConditionTimestamps[i] = c.Timestamp
	}
	for i := range firstPageConditions {
		// 新しいやつだけなら append
		if isNewData(targetIsu, &firstPageConditions[i]) {
			conditions = append(conditions, firstPageConditions[i])
		} else {
			// timestamp順なのは validation で保証しているので読んだやつが出てきたタイミングで return
			return conditions, newLastReadConditionTimestamps, nil
		}
	}
	// 最初のページが最後のページならやめる
	if len(firstPageConditions) < service.ConditionLimit {
		return conditions, newLastReadConditionTimestamps, nil
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

		requestTimeUnix = time.Now().Unix()
		tmpConditions, hres, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			return nil, newLastReadConditionTimestamps, []error{err}
		}
		err = verifyIsuConditions(hres, user, targetIsu.JIAIsuUUID, &request, tmpConditions, targetIsu.LastReadConditionTimestamps, requestTimeUnix)
		if err != nil {
			return nil, newLastReadConditionTimestamps, []error{err}
		}

		for i := range tmpConditions {
			// 新しいやつだけなら append
			if isNewData(targetIsu, &tmpConditions[i]) {
				conditions = append(conditions, tmpConditions[i])
			} else {
				// timestamp順なのは validation で保証しているので読んだやつが出てきたタイミングで return
				return conditions, newLastReadConditionTimestamps, nil
			}
		}

		// 最後のページまで見ちゃってるならやめる
		if len(tmpConditions) != service.ConditionLimit {
			return conditions, newLastReadConditionTimestamps, nil
		}
	}
}

func readCondition(conditions service.GetIsuConditionResponseArray, step *isucandar.BenchmarkStep, readConditionCount *ReadConditionCount) {
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
	return condition.Timestamp > isu.LastReadConditionTimestamps[0]
}

// あるISUのグラフを見に行くシナリオ
func (s *Scenario) requestGraphScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User, targetIsu *model.Isu, randEngine *rand.Rand) bool {
	// 最新の condition から、一度見た condition が帰ってくるまで condition のページングをする
	nowVirtualTime := s.ToVirtualTime(time.Now())
	// 割り算で切り捨てを発生させている(day単位にしている)
	virtualToday := trancateTimestampToDate(nowVirtualTime)
	virtualToday -= OneDay

	graphResponses, errs := getIsuGraphUntilLastViewed(ctx, user, targetIsu, virtualToday)
	if len(errs) > 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
		return false
	}

	// AddScoreはconditionのGETまで待つためここでタグを持っておく
	scoreTags := []score.ScoreTag{}

	// LastCompletedGraphTime を更新
	newLastCompletedGraphTime := getNewLastCompletedGraphTime(graphResponses, virtualToday)
	if targetIsu.LastCompletedGraphTime < newLastCompletedGraphTime {
		targetIsu.LastCompletedGraphTime = newLastCompletedGraphTime
	}

	// scoreの計算
	for behindDay, gr := range graphResponses {
		minTimestampCount := int(^uint(0) >> 1)
		// 「今日のグラフ」をリクエストした時刻が 01:00 より前のときのフラグ
		isTodayGraphOnly1Hour := false
		for hour, g := range *gr {
			// 「今日のグラフ」＆「リクエストした時間より先」ならもう minTimestampCount についてカウントしない
			if behindDay == 0 && nowVirtualTime.Unix() < g.EndAt {
				// 「今日のグラフ」をリクエストした時刻が 01:00 より前のとき
				if hour == 0 {
					isTodayGraphOnly1Hour = true
				}
				break
			}
			if len(g.ConditionTimestamps) < minTimestampCount {
				minTimestampCount = len(g.ConditionTimestamps)
			}
		}
		// 「今日のグラフ」をリクエストした時刻が 01:00 より前ならタグをつけずに次のループへ
		if isTodayGraphOnly1Hour {
			continue
		}
		// 「今日のグラフじゃない」＆「まだ見ていない完成しているグラフ」なら加点( graphResponses がまだ見ていないグラフの集合なのは保証されている)
		if behindDay != 0 && targetIsu.LastCompletedGraphTime >= virtualToday-(int64(behindDay)*OneDay) {
			// AddScoreはconditionのGETまで待つためここでタグを入れておく
			scoreTags = append(scoreTags, getGraphScoreTag(minTimestampCount))
		}
		// 「今日のグラフ」についても加点
		if behindDay == 0 {
			// AddScoreはconditionのGETまで待つためここでタグを入れておく
			scoreTags = append(scoreTags, getTodayGraphScoreTag(minTimestampCount))
		}
	}

	// graph の加点分を計算
	for _, scoreTag := range scoreTags {
		step.AddScore(scoreTag)
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
		requestTimeUnix := time.Now().Unix()
		conditions, hres, err := getIsuConditionAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return false
		}
		err = verifyIsuConditions(hres, user, targetIsu.JIAIsuUUID, &request, conditions, targetIsu.LastReadConditionTimestamps, requestTimeUnix)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return false
		}
	}

	return true
}

// unix timeのtimestampをその「日」に切り捨てる
func trancateTimestampToDate(now time.Time) int64 {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
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

func getTodayGraphScoreTag(minTimestampCount int) score.ScoreTag {
	if minTimestampCount > ScoreGraphTimestampCount.Good {
		return ScoreTodayGraphGood
	}
	if minTimestampCount > ScoreGraphTimestampCount.Normal {
		return ScoreTodayGraphNormal
	}
	if minTimestampCount > ScoreGraphTimestampCount.Bad {
		return ScoreTodayGraphBad
	}
	return ScoreTodayGraphWorst
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
	requestTimeUnix := time.Now().Unix()
	todayGraph, hres, err := getIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, todayRequest)
	if err != nil {
		return nil, []error{err}
	}
	err = verifyGraph(hres, user, targetIsu.JIAIsuUUID, &todayRequest, todayGraph, requestTimeUnix)
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
		requestTimeUnix = time.Now().Unix()

		tmpGraph, hres, err := getIsuGraphAction(ctx, user.Agent, targetIsu.JIAIsuUUID, request)
		if err != nil {
			return nil, []error{err}
		}
		err = verifyGraph(hres, user, targetIsu.JIAIsuUUID, &request, tmpGraph, requestTimeUnix)
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

func findBadIsuState(conditions service.GetIsuConditionResponseArray) (model.IsuStateChange, int64) {
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
		if bad && virtualTimestamp == 0 {
			virtualTimestamp = c.Timestamp
			break
		}
	}

	return solveCondition, virtualTimestamp
}

// 確率で signout して再度ログインするシナリオ。全てのシナリオの最後に確率で発生する
func signoutScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *model.User) {
	isSuccess := signoutInfinityRetry(ctx, user, step)
	if !isSuccess {
		return
	}

	user.Agent.ClearCookie()
	user.Agent.CacheStore.Clear()
	user.ClearStaticCache()

	// signout したらトップページに飛ぶ(MEMO: 初期状態だと trend おもすぎて backend をころしてしまうかも)
	isSuccess = browserGetLandingPageIgnoreInfinityRetry(ctx, user, step)
	if !isSuccess {
		return
	}

	authInfinityRetry(ctx, user, user.UserID, step)
}

// signoutScenario 以外からは呼ばない(シナリオループの最後である必要がある)
func authInfinityRetry(ctx context.Context, user *model.User, userID string, step *isucandar.BenchmarkStep) {
	for {
		select {
		case <-ctx.Done():
			// 失敗したときも区別せずに return してよい(次シナリオループで終了するため)
			return
		default:
		}
		_, errs := authAction(ctx, user, userID)
		if len(errs) > 0 {
			for _, err := range errs {
				addErrorWithContext(ctx, step, err)
			}
			continue
		}
		me, hres, err := getMeAction(ctx, user.GetAgent())
		if err != nil {
			addErrorWithContext(ctx, step, err)
			continue
		}
		err = verifyMe(userID, hres, me)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			continue
		}
		return
	}
}

func postIsuInfinityRetry(ctx context.Context, a *agent.Agent, req service.PostIsuRequest, step *isucandar.BenchmarkStep) (*service.Isu, *http.Response) {
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}
		isu, res, err := postIsuAction(ctx, a, req)
		if err != nil {
			if res != nil && res.StatusCode == http.StatusConflict {
				return nil, res
			}
			addErrorWithContext(ctx, step, err)
			continue
		}
		return isu, res
	}
}

// 返り値が false のときは常に ctx.Done
func signoutInfinityRetry(ctx context.Context, user *model.User, step *isucandar.BenchmarkStep) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		_, err := reqNoContentResNoContent(ctx, user.Agent, http.MethodPost, "/api/signout", []int{http.StatusOK, http.StatusUnauthorized})
		if err != nil {
			addErrorWithContext(ctx, step, err)
			continue
		}
		_, _, err = getMeErrorAction(ctx, user.Agent)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			continue
		}
		return true
	}
}

// 返り値が false のときは常に ctx.Done
func browserGetLandingPageIgnoreInfinityRetry(ctx context.Context, user *model.User, step *isucandar.BenchmarkStep) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		errs := browserGetLandingPageIgnoreAction(ctx, user)
		if len(errs) != 0 {
			for _, err := range errs {
				addErrorWithContext(ctx, step, err)
			}
			continue
		}
		return true
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
