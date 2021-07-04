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
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
)

//POST /api/isu/{jia_isu_id}/conditionをたたくスレッド
func KeepPosting(ctx context.Context, step *isucandar.BenchmarkStep, targetURL string, scenarioChan *model.StreamsForPoster) {
	defer func() { scenarioChan.ActiveChan <- false }() //deactivate

	type isuState struct {
		isDirty              bool
		lastClean            time.Time
		isOverweight         bool
		lastDetectOverweight time.Time
		isBroken             bool
		isuStateDelete       bool //椅子を削除する(正の点数が出るpostを行わない)
	}

	nowRealTime := time.Now()
	state := isuState{
		isDirty:              false,
		lastClean:            nowRealTime,
		isOverweight:         false,
		lastDetectOverweight: nowRealTime,
		isBroken:             false,
		isuStateDelete:       false,
	}
	randEngine := rand.New(rand.NewSource(0))

	//TODO: 頻度はちゃんと検討して変える
	//本来は1分=60,000msに一回
	//60倍速
	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		nowRealTime = time.Now()
		timeStamp := nowRealTime //TODO: 時間調整
		//乱数初期化（逆算できるように）
		randEngine.Seed(timeStamp.Unix() + 961054102)

		//状態変化
		stateChange := model.IsuStateChangeNone
		select {
		case nextState, ok := <-scenarioChan.StateChan:
			if ok {
				stateChange = nextState
			}
		default:
		}
		switch stateChange {
		case model.IsuStateChangeClear:
			state.isDirty = false
			state.lastClean = nowRealTime
		case model.IsuStateChangeDetectOverweight:
			state.isOverweight = false
			state.lastDetectOverweight = nowRealTime
		case model.IsuStateChangeClearAndDetect:
			state.isDirty = false
			state.lastClean = nowRealTime
			state.isOverweight = false
			state.lastDetectOverweight = nowRealTime
		case model.IsuStateChangeBad:
			randV := randEngine.Intn(100)
			if randV <= 70 {
				state.isDirty = true
			} else if randV <= 90 {
				state.isBroken = true
			} else {
				state.isDirty = true
				state.isBroken = true
			}
		case model.IsuStateChangeDelete:
			state.isuStateDelete = true
		}

		//TODO: 挙動をちゃんと変化させる

		condition := model.IsuCondition{
			IsSitting: (randEngine.Intn(100) <= 70),
			Condition: fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v",
				(randEngine.Intn(2) == 0),
				(randEngine.Intn(2) == 0),
				(randEngine.Intn(2) == 0),
			),
			Message:       "今日もいい天気",
			TimestampUnix: timeStamp.Unix(),
		}

		conditionByte, err := json.Marshal(condition)
		if err != nil {
			logger.AdminLogger.Panic(err)
			continue
		}

		func() {
			//TODO: 得点計算 step
			resp, err := http.Post(
				targetURL, "application/json",
				bytes.NewBuffer(conditionByte),
			)
			if err != nil {
				return // goto next loop
			}
			defer resp.Body.Close()

			if resp.StatusCode != 201 {
				return // goto next loop
			}
		}()
	}
}
