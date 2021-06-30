package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/labstack/gommon/log"
)

func KeepPosting(ctx context.Context, targetURL string, scenarioChan *model.IsuPosterChan) {
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
			log.Error(err)
			continue
		}

		func() {
			resp, err := http.Post(
				targetURL, "application/json",
				bytes.NewBuffer(conditionByte),
			)
			if err != nil {
				log.Error(err)
				return // goto next loop
			}
			defer resp.Body.Close()

			if resp.StatusCode != 201 {
				log.Errorf("failed to `POST %s` with status=`%s`", targetURL, resp.Status)
				return // goto next loop
			}
		}()
	}
}
