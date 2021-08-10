package scenario

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

const (
	// MEMO: 最大でも60秒に一件しか送れないので点数上限になるが、解決できるとは思えないので良い
	PostIntervalSecond     = 60 //Virtual Timeでのpost間隔
	PostIntervalBlurSecond = 5  //Virtual Timeでのpost間隔のブレ幅(+-PostIntervalBlurSecond)
	PostContentNum         = 10 //一回のpostで何要素postするか virtualTimeMulti * timerDuration(20ms) / PostIntervalSecond
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

//POST /api/condition/{jia_isu_id}をたたく Goroutine
func (s *Scenario) keepPosting(ctx context.Context, targetBaseURL *url.URL, isu *model.Isu, scenarioChan *model.StreamsForPoster) {
	postConditionTimeout := 100 * time.Millisecond //MEMO: timeout は気にせずにズバズバ投げる

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

		//TODO: 検証可能な生成方法にする
		//TODO: stateの適用タイミングをちゃんと考える

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
			stateChange = model.IsuStateChangeNone //TODO: stateの適用タイミングをちゃんと考える

			//リクエスト
			conditions = append(conditions, condition)
			conditionsReq = append(conditionsReq, service.PostIsuConditionRequest{
				IsSitting: condition.IsSitting,
				Condition: condition.ConditionString(),
				Message:   condition.Message,
				Timestamp: condition.TimestampUnix,
			})

		}

		if len(conditions) == 0 {
			continue
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
		// TODO: 70% なら 69 じゃない, 対して影響はない
		if randV <= 70 {
			state.dirty.isNow = true
		} else if randV <= 90 {
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

	// TODO: over_weight が true のときは sitting を false にしないように
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
