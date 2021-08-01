package scenario

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

const (
	PostInterval   = 5 * time.Minute //Virtual Timeでのpost間隔
	PostContentNum = 100             //一回のpostで何要素postするか
)

type posterState struct {
	//PostInterval         time.Duration
	lastCondition        model.IsuCondition
	lastClean            time.Time
	lastDetectOverweight time.Time
}

//POST /api/condition/{jia_isu_id}をたたく Goroutine
func (s *Scenario) keepPosting(ctx context.Context, step *isucandar.BenchmarkStep, targetBaseURL string, isu *model.Isu, scenarioChan *model.StreamsForPoster) {
	defer close(scenarioChan.ConditionChan)
	postConditionTimeout := 50 * time.Millisecond //MEMO: timeout は気にせずにズバズバ投げる

	nowTimeStamp := s.ToVirtualTime(time.Now())
	state := posterState{
		lastCondition: model.IsuCondition{
			IsSitting:     false,
			IsDirty:       false,
			IsOverweight:  false,
			IsBroken:      false,
			Message:       "",
			TimestampUnix: nowTimeStamp.Unix(),
		},
		lastClean:            nowTimeStamp,
		lastDetectOverweight: nowTimeStamp,
	}
	randEngine := rand.New(rand.NewSource(rand.Int63()))
	targetURL := fmt.Sprintf("%s/api/condition/%s", targetBaseURL, isu.JIAIsuUUID)
	httpClient := http.Client{}
	httpClient.Timeout = postConditionTimeout

	//post isuの待ち
	select {
	case <-ctx.Done():
		return
	case <-scenarioChan.StateChan:
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		nowTimeStamp = s.ToVirtualTime(time.Now())

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

		//TODO: 検証可能な生成方法にする
		//TODO: stateの適用タイミングをちゃんと考える
		conditions := []model.IsuCondition{}
		conditionsReq := []service.PostIsuConditionRequest{}
		for state.NextConditionTimestamp().Before(nowTimeStamp) {
			//次のstateを生成
			condition := state.GenerateNextCondition(randEngine, stateChange, isu) //TODO: stateの適用タイミングをちゃんと考える
			stateChange = model.IsuStateChangeNone                                 //TODO: stateの適用タイミングをちゃんと考える

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
		//postし損ねたconditionの数を制限
		if len(conditions) > PostContentNum {
			conditions = conditions[len(conditions)-PostContentNum:]
			conditionsReq = conditionsReq[len(conditionsReq)-PostContentNum:]
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

func (state *posterState) NextConditionTimestamp() time.Time {
	return time.Unix(state.lastCondition.TimestampUnix, 0).Add(PostInterval)
}
func (state *posterState) NextIsLatestTimestamp(nowTimeStamp time.Time) bool {
	return nowTimeStamp.Before(time.Unix(state.lastCondition.TimestampUnix, 0).Add(PostInterval * 2))
}
func (state *posterState) GenerateNextCondition(randEngine *rand.Rand, stateChange model.IsuStateChange, isu *model.Isu) model.IsuCondition {

	timeStamp := state.NextConditionTimestamp()

	//状態変化
	lastConditionIsDirty := state.lastCondition.IsDirty
	lastConditionIsOverweight := state.lastCondition.IsOverweight
	lastConditionIsBroken := state.lastCondition.IsBroken
	if stateChange == model.IsuStateChangeBad {
		randV := randEngine.Intn(100)
		// TODO: 70% なら 69 じゃない, 対して影響はない
		if randV <= 70 {
			lastConditionIsDirty = true
		} else if randV <= 90 {
			lastConditionIsBroken = true
		} else {
			lastConditionIsDirty = true
			lastConditionIsBroken = true
		}
	} else {
		//各種状態改善クエリ
		if stateChange&model.IsuStateChangeClear != 0 {
			lastConditionIsDirty = false
			state.lastClean = timeStamp
		}
		if stateChange&model.IsuStateChangeDetectOverweight != 0 {
			lastConditionIsDirty = false
			state.lastClean = timeStamp
		}
		if stateChange&model.IsuStateChangeRepair != 0 {
			lastConditionIsBroken = false
		}
	}

	//新しいConditionを生成
	condition := model.IsuCondition{
		StateChange:  stateChange,
		IsSitting:    state.lastCondition.IsSitting,
		IsDirty:      lastConditionIsDirty,
		IsOverweight: lastConditionIsOverweight,
		IsBroken:     lastConditionIsBroken,
		//ConditionLevel: model.ConditionLevelCritical,
		Message:       "",
		TimestampUnix: timeStamp.Unix(),
		OwnerIsuUUID:  isu.JIAIsuUUID,
		OwnerIsuID:    isu.ID,
	}
	// TODO: over_weight が true のときは sitting を false にしないように
	//sitting
	if condition.IsSitting {
		if randEngine.Intn(100) <= 10 {
			condition.IsSitting = false
			condition.IsOverweight = false
		}
	} else {
		if randEngine.Intn(100) <= 10 {
			condition.IsSitting = true
		}
	}
	//overweight
	if condition.IsSitting && timeStamp.Sub(state.lastDetectOverweight) > 60*time.Minute {
		if randEngine.Intn(100) <= 5 {
			condition.IsOverweight = true
		}
	}
	//dirty
	if timeStamp.Sub(state.lastClean) > 75*time.Minute {
		if randEngine.Intn(100) <= 5 {
			condition.IsDirty = true
		}
	}
	//broken
	if randEngine.Intn(1000) <= 1 {
		condition.IsBroken = true
	}

	//message
	condition.Message = "今日もいい天気" //TODO: メッセージをちゃんと生成

	//conditionLevel
	condition.ConditionLevel = calcConditionLevel(condition)

	//last更新
	state.lastCondition = condition

	return condition
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
