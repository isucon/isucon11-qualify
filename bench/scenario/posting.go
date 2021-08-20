package scenario

import (
	"context"
	"crypto/tls"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

const (
	// MEMO: 最大でも60秒に一件しか送れないので点数上限になるが、解決できるとは思えないので良い
	PostIntervalSecond     = 60                     //Virtual Timeでのpost間隔
	PostIntervalBlurSecond = 5                      //Virtual Timeでのpost間隔のブレ幅(+-PostIntervalBlurSecond)
	PostContentNum         = 10                     //一回のpostで何要素postするか virtualTimeMulti * timerDuration(20ms) / PostIntervalSecond
	postConditionTimeout   = 100 * time.Millisecond //MEMO: timeout は気にせずにズバズバ投げる
)

var (
	targetBaseURLMapMutex sync.Mutex
	targetBaseURLMap      = map[string]string{}
)

var (
	posterWaitGroup sync.WaitGroup
	// 全ユーザーがpostしたconditionの端数の合計。Goroutine終了時に加算する
	postInfoConditionFraction     int32 = 0
	postWarnConditionFraction     int32 = 0
	postCriticalConditionFraction int32 = 0
)

func init() {
	if !(2*PostIntervalBlurSecond < PostIntervalSecond) {
		panic("assert: 2*PostIntervalBlurSecond < PostIntervalSecond")
	}
}

type posterState struct {
	lastConditionTimestamp int64
	isSitting              bool
	dirty                  badCondition
	overWeight             badCondition
	broken                 badCondition
}

type badCondition struct {
	fixedTime int64
	isNow     bool
}

func (s *Scenario) postConditionNumReporter(ctx context.Context, step *isucandar.BenchmarkStep) {
	var postInfoConditionNum int32 = 0
	var postWarnConditionNum int32 = 0
	var postCriticalConditionNum int32 = 0
	addScore := func() {
		postInfoConditionNum += atomic.SwapInt32(&postInfoConditionFraction, 0)
		postWarnConditionNum += atomic.SwapInt32(&postWarnConditionFraction, 0)
		postCriticalConditionNum += atomic.SwapInt32(&postCriticalConditionFraction, 0)

		for postInfoConditionNum > ReadConditionTagStep {
			postInfoConditionNum -= ReadConditionTagStep
			step.AddScore(ScorePostInfoCondition)
		}
		for postWarnConditionNum > ReadConditionTagStep {
			postWarnConditionNum -= ReadConditionTagStep
			step.AddScore(ScorePostWarningCondition)
		}
		for postCriticalConditionNum > ReadConditionTagStep {
			postCriticalConditionNum -= ReadConditionTagStep
			step.AddScore(ScorePostCriticalCondition)
		}
	}
	for {
		time.Sleep(1500 * time.Millisecond)

		addScore()
		select {
		case <-ctx.Done():
			posterWaitGroup.Wait()
			addScore()
			return
		default:
		}
	}
}

//POST /api/condition/{jia_isu_id}をたたく Goroutine
func (s *Scenario) keepPosting(ctx context.Context, targetBaseURL *url.URL, fqdn string, isu *model.Isu, scenarioChan *model.StreamsForPoster) {

	targetBaseURLMapMutex.Lock()
	targetBaseURLMap[targetBaseURL.String()] = fqdn
	targetBaseURLMapMutex.Unlock()

	posterWaitGroup.Add(1)
	var postInfoConditionNum int32 = 0
	var postWarnConditionNum int32 = 0
	var postCriticalConditionNum int32 = 0
	defer func() {
		atomic.AddInt32(&postInfoConditionFraction, postInfoConditionNum)
		atomic.AddInt32(&postWarnConditionFraction, postWarnConditionNum)
		atomic.AddInt32(&postCriticalConditionFraction, postCriticalConditionNum)
		posterWaitGroup.Done()
	}()

	targetBaseURL.Path = path.Join(targetBaseURL.Path, "/api/condition/", isu.JIAIsuUUID)
	nowTimeStamp := s.ToVirtualTime(time.Now()).Unix()
	state := posterState{
		// lastConditionTimestamp: 0,
		lastConditionTimestamp: nowTimeStamp,
		dirty:                  badCondition{0, false},
		overWeight:             badCondition{0, false},
		broken:                 badCondition{0, false},
		isSitting:              false,
	}
	randEngine := rand.New(rand.NewSource(rand.Int63()))
	httpClient := http.Client{}
	httpClient.Timeout = postConditionTimeout
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: fqdn,
		},
		ForceAttemptHTTP2: true,
	}

	timer := time.NewTicker(40 * time.Millisecond)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		nowTimeStamp = s.ToVirtualTime(time.Now()).Unix()

		//状態変化
		stateChange := model.IsuStateChangeNone
		select {
		case nextState, ok := <-scenarioChan.StateChan:
			if ok {
				stateChange = nextState
			} else {
				// StateChan が閉じられるときは load が終了したとき
				return
			}
		default:
		}

		// 今の時間から最後のconditionの時間を引く
		diffTimestamp := nowTimeStamp - state.lastConditionTimestamp
		// その間に何個のconditionがあったか
		diffConditionCount := int((diffTimestamp + PostIntervalSecond - 1) / PostIntervalSecond)

		var reqLength int
		if diffConditionCount > PostContentNum {
			reqLength = PostContentNum
		} else {
			reqLength = diffConditionCount
		}

		conditions := make([]model.IsuCondition, 0, reqLength)
		conditionsReq := make([]service.PostIsuConditionRequest, 0, reqLength)

		for i := 0; i < diffConditionCount; i++ {
			// 次のstateを生成
			state.UpdateToNextState(randEngine, stateChange)

			if i+PostContentNum-diffConditionCount < 0 {
				continue
			}

			// 作った新しいstateに基づいてconditionを生成
			condition := state.GetNewestCondition(randEngine, stateChange, isu)
			stateChange = model.IsuStateChangeNone

			//リクエスト
			conditions = append(conditions, condition)
			conditionsReq = append(conditionsReq, service.PostIsuConditionRequest{
				IsSitting: condition.IsSitting,
				Condition: condition.ConditionString(),
				Message:   condition.Message,
				Timestamp: condition.TimestampUnix,
			})

			switch condition.ConditionLevel {
			case model.ConditionLevelInfo:
				postInfoConditionNum++
			case model.ConditionLevelWarning:
				postWarnConditionNum++
			case model.ConditionLevelCritical:
				postCriticalConditionNum++
			}
		}

		if len(conditions) == 0 {
			continue
		}

		if postInfoConditionNum > ReadConditionTagStep {
			atomic.AddInt32(&postInfoConditionFraction, postInfoConditionNum)
			postInfoConditionNum = 0
		}
		if postWarnConditionNum > ReadConditionTagStep {
			atomic.AddInt32(&postWarnConditionFraction, postWarnConditionNum)
			postWarnConditionNum = 0
		}
		if postCriticalConditionNum > ReadConditionTagStep {
			atomic.AddInt32(&postCriticalConditionFraction, postCriticalConditionNum)
			postCriticalConditionNum = 0
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
		isu.AddIsuConditions(conditions)

		// timeout も無視するので全てのエラーを見ない
		postIsuConditionAction(ctx, httpClient, targetBaseURL.String(), &conditionsReq)
	}
}

func (state *posterState) NextConditionTimeStamp() int64 {
	return state.lastConditionTimestamp + PostIntervalSecond
}

func (state *posterState) GetNewestCondition(randEngine *rand.Rand, stateChange model.IsuStateChange, isu *model.Isu) model.IsuCondition {

	// ハック対策に PostIntervalSecond にずれを出してる
	blur := randEngine.Int63n(2*PostIntervalBlurSecond+1) - PostIntervalBlurSecond
	//新しいConditionを生成
	condition := model.IsuCondition{
		StateChange:  stateChange,
		IsSitting:    state.isSitting,
		IsDirty:      state.dirty.isNow,
		IsOverweight: state.overWeight.isNow,
		IsBroken:     state.broken.isNow,
		//ConditionLevel: model.ConditionLevelCritical,
		Message:       "",
		TimestampUnix: state.lastConditionTimestamp + blur,
		ReadTime:      math.MaxInt64 - ConditionDelayTime, // 減算しているのはオーバーフロー対策
	}

	//message
	condition.Message = random.MessageWithCondition(state.dirty.isNow, state.overWeight.isNow, state.broken.isNow, isu.CharacterID)

	//conditionLevel
	condition.ConditionLevel = calcConditionLevel(condition)

	return condition
}

func (state *posterState) UpdateToNextState(randEngine *rand.Rand, stateChange model.IsuStateChange) {

	timeStamp := state.NextConditionTimeStamp()
	state.lastConditionTimestamp = timeStamp

	//状態変化
	if stateChange == model.IsuStateChangeBad {
		randV := randEngine.Intn(100)
		if randV < 70 {
			state.dirty.isNow = true
		} else if randV < 90 {
			state.broken.isNow = true
		} else {
			state.dirty.isNow = true
			state.broken.isNow = true
		}
	} else {
		//各種状態改善クエリ
		if stateChange&model.IsuStateChangeClear != 0 {
			state.dirty.isNow = false
			state.dirty.fixedTime = timeStamp
		}
		if stateChange&model.IsuStateChangeDetectOverweight != 0 {
			state.overWeight.isNow = false
			state.overWeight.fixedTime = timeStamp
		}
		if stateChange&model.IsuStateChangeRepair != 0 {
			state.broken.isNow = false
			state.broken.fixedTime = timeStamp
		}
	}

	//sitting
	if state.isSitting {
		// sitting が false になるのは over_weight が true じゃないとき
		if !state.overWeight.isNow {
			if randEngine.Intn(100) <= 10 {
				state.isSitting = false
			}
		}
	} else {
		if randEngine.Intn(100) <= 10 {
			state.isSitting = true
		}
	}
	//overweight
	if state.isSitting && timeStamp-state.overWeight.fixedTime > 12*60*60 {
		if randEngine.Intn(5000) <= 1 {
			state.overWeight.isNow = true
		}
	}
	//dirty
	if timeStamp-state.dirty.fixedTime > 18*60*60 {
		if randEngine.Intn(5000) <= 1 {
			state.dirty.isNow = true
		}
	}
	//broken
	if timeStamp-state.broken.fixedTime > 24*60*60 {
		if randEngine.Intn(10000) <= 1 {
			state.broken.isNow = true
		}
	}
}

func calcConditionLevel(condition model.IsuCondition) model.ConditionLevel {
	warnCount := 0
	if condition.IsDirty {
		warnCount += 1
	}
	if condition.IsOverweight {
		warnCount += 1
	}
	if condition.IsBroken {
		warnCount += 1
	}
	if warnCount == 0 {
		return model.ConditionLevelInfo
	} else if warnCount == 1 || warnCount == 2 {
		return model.ConditionLevelWarning
	} else {
		return model.ConditionLevelCritical
	}
}

//invalid

//ランダムなISUにconditionを投げる
func (s *Scenario) keepPostingError(ctx context.Context) {
	nowTimeStamp := s.ToVirtualTime(time.Now()).Unix()
	state := posterState{
		// lastConditionTimestamp: 0,
		lastConditionTimestamp: nowTimeStamp,
		dirty:                  badCondition{0, false},
		overWeight:             badCondition{0, false},
		broken:                 badCondition{0, false},
		isSitting:              false,
	}
	randEngine := rand.New(rand.NewSource(rand.Int63()))
	httpClient := http.Client{}
	httpClient.Timeout = postConditionTimeout
	httpClient.Transport = &http.Transport{
		TLSClientConfig:   &tls.Config{},
		ForceAttemptHTTP2: true,
	}

	timer := time.NewTicker(1000 * time.Millisecond)
	defer timer.Stop()
	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		count++

		nowTimeStamp = s.ToVirtualTime(time.Now()).Unix()

		//ISUを選ぶ
		isu := s.GetRandomActivatedIsu(randEngine)
		if isu == nil {
			continue
		}

		//状態変化
		stateChange := model.IsuStateChangeNone
		if count%5 == 0 {
			stateChange = model.IsuStateChangeClear | model.IsuStateChangeDetectOverweight | model.IsuStateChangeRepair
		}

		// 今の時間から最後のconditionの時間を引く
		diffTimestamp := nowTimeStamp - state.lastConditionTimestamp
		// その間に何個のconditionがあったか
		diffConditionCount := int((diffTimestamp + PostIntervalSecond - 1) / PostIntervalSecond)

		var reqLength int
		if diffConditionCount > PostContentNum {
			reqLength = PostContentNum
		} else {
			reqLength = diffConditionCount
		}
		conditionsReq := make([]service.PostIsuConditionRequest, 0, reqLength)

		for i := 0; i < diffConditionCount; i++ {
			// 次のstateを生成
			state.UpdateToNextState(randEngine, stateChange)

			if i+PostContentNum-diffConditionCount < 0 {
				continue
			}

			// 作った新しいstateに基づいてconditionを生成
			condition := state.GetNewestCondition(randEngine, stateChange, isu)
			stateChange = model.IsuStateChangeNone //TODO: stateの適用タイミングをちゃんと考える

			data := service.PostIsuConditionRequest{
				IsSitting: condition.IsSitting,
				Condition: condition.ConditionString(),
				Message:   condition.Message,
				Timestamp: condition.TimestampUnix,
			}

			//Conditionのフォーマットを崩す
			index := randEngine.Intn(len(data.Condition) - 1)
			data.Condition = data.Condition[:index] +
				string((data.Condition[index]-'a'+byte(randEngine.Intn(26)))%26+'a') +
				data.Condition[index+1:]

			//リクエスト
			conditionsReq = append(conditionsReq, data)
		}

		if len(conditionsReq) == 0 {
			continue
		}

		//必ず一つは間違っているようにする
		if randEngine.Intn(2) == 0 {
			conditionsReq[len(conditionsReq)/2].Condition = "is_dirty=true,is_overweight=true,is_brokan=false"
		} else {
			conditionsReq[len(conditionsReq)/2].Condition = "is_dirty:true,is_overweight:true,is_broken:false"
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		// targetを取得
		var targetPath string
		var targetServer string
		targetBaseURLMapMutex.Lock()
		for pathTmp, serverTmp := range targetBaseURLMap {
			targetPath = pathTmp
			targetServer = serverTmp
			break
		}
		targetBaseURLMapMutex.Unlock()
		targetPath = path.Join(targetPath, "/api/condition/", isu.JIAIsuUUID)
		httpClient.Transport.(*http.Transport).TLSClientConfig.ServerName = targetServer
		// timeout も無視するので全てのエラーを見ない
		postIsuConditionAction(ctx, httpClient, targetPath, &conditionsReq)
	}
}
