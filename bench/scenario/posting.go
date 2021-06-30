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
func KeepPosting(ctx context.Context, step *isucandar.BenchmarkStep, targetURL string, scenarioChan *model.IsuPosterChan) {
	randEngine := rand.New(rand.NewSource(0))

	timer := time.NewTicker(2 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		//TODO: 挙動をちゃんと変化させる
		//TODO: 時間加速

		//乱数初期化（逆算できるように）
		nowTime := time.Now()
		randEngine.Seed(nowTime.UnixNano()/1000000000 + 961054102)

		condition := model.IsuCondition{
			IsSitting: (randEngine.Intn(100) <= 70),
			Condition: fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v",
				(randEngine.Intn(2) == 0),
				(randEngine.Intn(2) == 0),
				(randEngine.Intn(2) == 0),
			),
			Message:       "今日もいい天気",
			TimestampUnix: nowTime.Unix(),
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
