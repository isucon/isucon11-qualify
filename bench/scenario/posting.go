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
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

const (
	PostInterval      = 5 * time.Minute //Virtual Timeでのpost間隔
	PostContentNum    = 100             //一回のpostで何要素postするか
	ConditionTagCount = 100             //condition 100件ごとに1タグ
)

type posterState struct {
	//PostInterval         time.Duration
	lastCondition        model.IsuCondition
	lastClean            time.Time
	lastDetectOverweight time.Time
	isuStateDelete       bool //椅子を削除する(正の点数が出るpostを行わない)
}

//POST /api/isu/{jia_isu_id}/conditionをたたく Goroutine
func (s *Scenario) keepPosting(ctx context.Context, step *isucandar.BenchmarkStep, targetBaseURL string, jiaIsuUUID string, scenarioChan *model.StreamsForPoster) {
	defer func() { scenarioChan.ActiveChan <- false }() //deactivate 容量1で、ここでしか使わないのでブロックしない

	userTimer, userTimerCancel := context.WithDeadline(ctx, s.realTimeLoadFinishedAt.Add(-agent.DefaultRequestTimeout-1*time.Second))
	defer userTimerCancel()

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
	httpClient.Timeout = agent.DefaultRequestTimeout + 5*time.Second //MEMO: post conditionがtimeoutすると付随してたくさんエラーが出るので、timeoutしにくいようにする

	conditionInfoCount := 0
	conditionWarningCount := 0
	conditionCriticalCount := 0

	//post isuの待ち
	select {
	case <-ctx.Done():
		return
	case <-userTimer.Done():
		return
	case stateChange := <-scenarioChan.StateChan:
		if stateChange == model.IsuStateChangeDelete {
			return
		}
	}

	//TODO: 頻度はちゃんと検討して変える
	timer := time.NewTicker(PostInterval * PostContentNum / s.virtualTimeMulti)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-userTimer.Done():
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

		//TODO: 検証可能な生成方法にする
		//TODO: stateの適用タイミングをちゃんと考える
		conditions := []model.IsuCondition{}
		conditionsReq := []service.PostIsuConditionRequest{}
		for state.NextConditionTimestamp().Before(nowTimeStamp) {
			//次のstateを生成
			condition := state.GenerateNextCondition(randEngine, stateChange, jiaIsuUUID) //TODO: stateの適用タイミングをちゃんと考える
			stateChange = model.IsuStateChangeNone                                        //TODO: stateの適用タイミングをちゃんと考える

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

		conditionByte, err := json.Marshal(conditionsReq)
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
				addErrorWithContext(ctx, step, err)
				return // goto next loop
			}

			for _, condition := range conditions {
				switch condition.ConditionLevel {
				case model.ConditionLevelInfo:
					conditionInfoCount++
				case model.ConditionLevelWarning:
					conditionWarningCount++
				case model.ConditionLevelCritical:
					conditionCriticalCount++
				}
			}

			//TODO: ユーザー Goroutineが詰まると詰まるのでいや
			select {
			case <-ctx.Done():
				return
			case scenarioChan.ConditionChan <- conditions:
			}

			//TODO: スレッドを終了する際に、端数をグローバル変数に逃がして、後で集計できるようにする
			for conditionInfoCount >= ConditionTagCount {
				step.AddScore(ScorePostConditionInfo)
				conditionInfoCount -= ConditionTagCount
			}
			for conditionWarningCount >= ConditionTagCount {
				step.AddScore(ScorePostConditionWarning)
				conditionWarningCount -= ConditionTagCount
			}
			for conditionCriticalCount >= ConditionTagCount {
				step.AddScore(ScorePostConditionCritical)
				conditionCriticalCount -= ConditionTagCount
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
func (state *posterState) GenerateNextCondition(randEngine *rand.Rand, stateChange model.IsuStateChange, jiaIsuUUID string) model.IsuCondition {

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
	var condition model.IsuCondition
	if state.isuStateDelete {
		//削除された椅子のConditionは0点固定
		condition = model.IsuCondition{
			StateChange:    model.IsuStateChangeDelete,
			IsSitting:      true,
			IsDirty:        true,
			IsOverweight:   true,
			IsBroken:       true,
			ConditionLevel: model.ConditionLevelCritical,
			Message:        "",
			TimestampUnix:  timeStamp.Unix(),
			OwnerID:        jiaIsuUUID,
		}
	} else {
		//新しいConditionを生成
		condition = model.IsuCondition{
			StateChange:  stateChange,
			IsSitting:    state.lastCondition.IsSitting,
			IsDirty:      lastConditionIsDirty,
			IsOverweight: lastConditionIsOverweight,
			IsBroken:     lastConditionIsBroken,
			//ConditionLevel: model.ConditionLevelCritical,
			Message:       "",
			TimestampUnix: timeStamp.Unix(),
			OwnerID:       jiaIsuUUID,
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
	state.lastCondition = condition

	return condition
}
