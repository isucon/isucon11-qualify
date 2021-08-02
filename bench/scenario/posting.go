package scenario

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

const (
	// MEMO: 最大でも一秒に一件しか送れないので点数上限になるが、解決できるとは思えないので良い
	PostIntervalSecond = 60  //Virtual Timeでのpost間隔
	PostContentNum     = 100 //一回のpostで何要素postするか
)

type posterState struct {
	lastConditionTimestamp        int64
	lastCleanTimestamp            int64
	lastDetectOverweightTimestamp int64
	lastRepairTimestamp           int64
	lastConditionIsSitting        bool
	lastConditionIsDirty          bool
	lastConditionIsBroken         bool
	lastConditionIsOverweight     bool
}

//POST /api/condition/{jia_isu_id}をたたく Goroutine
func (s *Scenario) keepPosting(ctx context.Context, step *isucandar.BenchmarkStep, targetBaseURL string, isu *model.Isu, scenarioChan *model.StreamsForPoster) {
	defer close(scenarioChan.ConditionChan)
	postConditionTimeout := 50 * time.Millisecond //MEMO: timeout は気にせずにズバズバ投げる

	nowTime := s.ToVirtualTime(time.Now())
	state := posterState{
		// lastConditionTimestamp: 0,
		lastConditionTimestamp:        nowTime.Unix(),
		lastCleanTimestamp:            0,
		lastDetectOverweightTimestamp: 0,
		lastRepairTimestamp:           0,
		lastConditionIsSitting:        false,
		lastConditionIsDirty:          false,
		lastConditionIsBroken:         false,
		lastConditionIsOverweight:     false,
	}
	randEngine := rand.New(rand.NewSource(rand.Int63()))
	targetURL := fmt.Sprintf("%s/api/condition/%s", targetBaseURL, isu.JIAIsuUUID)
	httpClient := http.Client{}
	httpClient.Timeout = postConditionTimeout

	timer := time.NewTicker(20 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		nowTimeStamp := s.ToVirtualTime(time.Now()).Unix()

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
		diffConditionCount := int(math.Ceil(float64(diffTimestamp) / float64(PostIntervalSecond)))

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
			condition := state.GetNewestCondition(stateChange, isu)
			stateChange = model.IsuStateChangeNone //TODO: stateの適用タイミングをちゃんと考える

			//リクエスト
			conditions = append(conditions, condition)
			conditionsReq = append(conditionsReq, service.PostIsuConditionRequest{
				IsSitting: condition.IsSitting,
				Condition: fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v",
					condition.IsDirty,
					condition.IsOverweight,
					condition.IsBroken,
				),
				Message:   condition.Message,
				Timestamp: condition.TimestampUnix,
			})
		}

		if len(conditions) == 0 {
			continue
		}

		//TODO: ユーザー Goroutineが詰まると詰まるのでいや
		select {
		case <-ctx.Done():
			return
		case scenarioChan.ConditionChan <- conditions:
		}

		// timeout も無視するので全てのエラーを見ない
		postIsuConditionAction(httpClient, targetURL, &conditionsReq)
	}
}

func (state *posterState) NextConditionTimeStamp() int64 {
	return state.lastConditionTimestamp + PostIntervalSecond
}

func (state *posterState) GetNewestCondition(stateChange model.IsuStateChange, isu *model.Isu) model.IsuCondition {

	//新しいConditionを生成
	condition := model.IsuCondition{
		StateChange:  stateChange,
		IsSitting:    state.lastConditionIsSitting,
		IsDirty:      state.lastConditionIsDirty,
		IsOverweight: state.lastConditionIsOverweight,
		IsBroken:     state.lastConditionIsBroken,
		//ConditionLevel: model.ConditionLevelCritical,
		Message:       "",
		TimestampUnix: state.lastConditionTimestamp,
		OwnerIsuUUID:  isu.JIAIsuUUID,
		OwnerIsuID:    isu.ID,
	}

	//message
	condition.Message = "今日もいい天気" //TODO: メッセージをちゃんと生成

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
			state.lastConditionIsDirty = true
		} else if randV <= 90 {
			state.lastConditionIsBroken = true
		} else {
			state.lastConditionIsDirty = true
			state.lastConditionIsBroken = true
		}
	} else {
		//各種状態改善クエリ
		if stateChange&model.IsuStateChangeClear != 0 {
			state.lastConditionIsDirty = false
			state.lastCleanTimestamp = timeStamp
		}
		if stateChange&model.IsuStateChangeDetectOverweight != 0 {
			state.lastConditionIsDirty = false
			state.lastCleanTimestamp = timeStamp
		}
		if stateChange&model.IsuStateChangeRepair != 0 {
			state.lastConditionIsBroken = false
			state.lastRepairTimestamp = timeStamp
		}
	}

	// TODO: over_weight が true のときは sitting を false にしないように
	//sitting
	if state.lastConditionIsSitting {
		// sitting が false になるのは over_weight が true じゃないとき
		if !state.lastConditionIsOverweight {
			if randEngine.Intn(100) <= 10 {
				state.lastConditionIsSitting = false
			}
		}
	} else {
		if randEngine.Intn(100) <= 10 {
			state.lastConditionIsSitting = true
		}
	}
	//overweight
	if state.lastConditionIsSitting && timeStamp-state.lastDetectOverweightTimestamp > 60*60 {
		if randEngine.Intn(100) <= 5 {
			state.lastConditionIsOverweight = true
		}
	}
	//dirty
	if timeStamp-state.lastCleanTimestamp > 75*60 {
		if randEngine.Intn(100) <= 5 {
			state.lastConditionIsDirty = true
		}
	}
	//broken
	if timeStamp-state.lastRepairTimestamp > 120*60 {
		state.lastConditionIsBroken = true
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
