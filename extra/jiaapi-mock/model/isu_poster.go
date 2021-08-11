package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
)

type IsuConditionPoster struct {
	TargetURL string
	IsuUUID   string

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewIsuConditionPoster(targetURL string, isuUUID string) IsuConditionPoster {
	ctx, cancel := context.WithCancel(context.Background())
	return IsuConditionPoster{targetURL, isuUUID, ctx, cancel}
}

func (m *IsuConditionPoster) KeepPosting() {
	targetURL := fmt.Sprintf(
		"%s/api/condition/%s",
		m.TargetURL, m.IsuUUID,
	)
	randEngine := rand.New(rand.NewSource(0))

	timer := time.NewTicker(2 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-timer.C:
		}

		//乱数初期化（逆算できるように）
		nowTime := time.Now()
		randEngine.Seed(nowTime.UnixNano()/1000000000 + 961054102)

		notification, err := json.Marshal(IsuNotificationRequest{{
			IsSitting: (randEngine.Intn(100) <= 70),
			Condition: fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v",
				(randEngine.Intn(2) == 0),
				(randEngine.Intn(2) == 0),
				(randEngine.Intn(2) == 0),
			),
			Message:   "テストメッセージです",
			Timestamp: nowTime.Unix(),
		}})
		if err != nil {
			log.Error(err)
			continue
		}

		func() {
			resp, err := http.Post(
				targetURL, "application/json",
				bytes.NewBuffer(notification),
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

func (m *IsuConditionPoster) StopPosting() {
	m.cancelFunc()
}

type IsuNotificationRequest []struct {
	IsSitting bool   `json:"is_sitting"`
	Condition string `json:"condition"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}
