package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
)

const (
	PostInterval   = 5 * time.Minute //Virtual Timeでのpost間隔
	PostContentNum = 10              //一回のpostで何要素postするか
)

type posterState struct {
	//PostInterval         time.Duration
	lastCondition        model.IsuCondition
	lastClean            time.Time
	lastDetectOverweight time.Time
	isuStateDelete       bool //椅子を削除する(正の点数が出るpostを行わない)
}

//POST /api/isu/{jia_isu_id}/conditionをたたくスレッド
func (s *Scenario) keepPosting(ctx context.Context, step *isucandar.BenchmarkStep, targetBaseURL string, jiaIsuUUID string, scenarioChan *model.StreamsForPoster) {
	defer func() { scenarioChan.ActiveChan <- false }() //deactivate

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
		isuStateDelete:       false,
	}
	randEngine := rand.New(rand.NewSource(0))
	targetURL := fmt.Sprintf("%s/api/isu/%s/condition", targetBaseURL, jiaIsuUUID)
	httpClient := http.Client{}
	httpClient.Timeout = 1 * time.Second

	//TODO: 頻度はちゃんと検討して変える
	timer := time.NewTicker(PostInterval * PostContentNum / s.virtualTimeMulti)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		nowTimeStamp = s.ToVirtualTime(time.Now())

		//状態変化
		stateChange := model.IsuStateChangeNone
		select {
		case nextState, ok := <-scenarioChan.StateChan:
			if ok {
				stateChange = nextState
			}
		default:
		}

		//TODO: まとめて投げる

		//postし損ねたconditionを捨てる
		for !state.NextIsLatestTimestamp(nowTimeStamp) {
			_ = state.GenerateNextCondition(randEngine, model.IsuStateChangeNone)
		}
		//次のstateを生成
		condition := state.GenerateNextCondition(randEngine, stateChange)

		//リクエスト
		conditionByte, err := json.Marshal(condition)
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		res, err := httpClient.Post(
			targetURL, "application/json",
			bytes.NewBuffer(conditionByte),
		)
		if err != nil {
			step.AddError(failure.NewError(ErrHTTP, err))
			continue // goto next loop
		}
		func() {
			defer res.Body.Close()

			err = verifyStatusCode(res, http.StatusCreated)
			if err != nil {
				step.AddError(err)
				return // goto next loop
			}

			if condition.ConditionLevel == model.ConditionLevelInfo {
				step.AddScore(ScorePostConditionInfo)
			} else if condition.ConditionLevel == model.ConditionLevelWarning {
				step.AddScore(ScorePostConditionWarning)
			} else {
				step.AddScore(ScorePostConditionCritical)
			}
		}()
	}
}

func (state *posterState) NextConditionTimestamp() time.Time {
	return time.Unix(state.lastCondition.TimestampUnix, 0).Add(PostInterval)
}
func (state *posterState) NextIsLatestTimestamp(nowTimeStamp time.Time) bool {
	return nowTimeStamp.Before(time.Unix(state.lastCondition.TimestampUnix, 0).Add(PostInterval * 2))
}
func (state *posterState) GenerateNextCondition(randEngine *rand.Rand, stateChange model.IsuStateChange) *model.IsuCondition {

	//乱数初期化（逆算できるように）
	timeStamp := state.NextConditionTimestamp()
	randEngine.Seed(timeStamp.Unix() + 961054102)

	//状態変化
	lastConditionIsDirty := state.lastCondition.IsDirty
	lastConditionIsOverweight := state.lastCondition.IsOverweight
	lastConditionIsBroken := state.lastCondition.IsBroken
	if stateChange == model.IsuStateChangeBad {
		randV := randEngine.Intn(100)
		if randV <= 70 {
			lastConditionIsDirty = true
		} else if randV <= 90 {
			lastConditionIsBroken = true
		} else {
			lastConditionIsDirty = true
			lastConditionIsBroken = true
		}
	} else if stateChange == model.IsuStateChangeDelete {
		state.isuStateDelete = true
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
	var condition *model.IsuCondition
	if state.isuStateDelete {
		//削除された椅子のConditionは0点固定
		condition = &model.IsuCondition{
			StateChange:    model.IsuStateChangeDelete,
			IsSitting:      true,
			IsDirty:        true,
			IsOverweight:   true,
			IsBroken:       true,
			ConditionLevel: model.ConditionLevelCritical,
			Message:        "",
			TimestampUnix:  timeStamp.Unix(),
		}
	} else {
		//新しいConditionを生成
		condition = &model.IsuCondition{
			StateChange:  stateChange,
			IsSitting:    state.lastCondition.IsSitting,
			IsDirty:      lastConditionIsDirty,
			IsOverweight: lastConditionIsOverweight,
			IsBroken:     lastConditionIsBroken,
			//ConditionLevel: model.ConditionLevelCritical,
			Message:       "",
			TimestampUnix: timeStamp.Unix(),
		}
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
			if randEngine.Intn(100) <= 50 {
				condition.IsOverweight = true
			}
		}
		//dirty
		if timeStamp.Sub(state.lastClean) > 75*time.Minute {
			if randEngine.Intn(100) <= 50 {
				condition.IsDirty = true
			}
		}
		//broken
		if randEngine.Intn(100) <= 1 {
			condition.IsBroken = true
		}

		//message
		condition.Message = "今日もいい天気" //TODO: メッセージをちゃんと生成

		//conditionLevel
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
			condition.ConditionLevel = model.ConditionLevelInfo
		} else if warnCount == 1 || warnCount == 2 {
			condition.ConditionLevel = model.ConditionLevelWarning
		} else {
			condition.ConditionLevel = model.ConditionLevelCritical
		}
	}

	//last更新
	state.lastCondition = *condition

	return condition
}
