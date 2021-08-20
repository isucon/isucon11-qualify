package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/labstack/gommon/log"
)

const postingIntervalSec = 3

type IsuConditionPoster struct {
	TargetURL url.URL
	IsuUUID   string

	ctx        context.Context
	cancelFunc context.CancelFunc
}

type IsuConditionRequest struct {
	IsSitting bool   `json:"is_sitting"`
	Condition string `json:"condition"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func NewIsuConditionPoster(targetURL *url.URL, isuUUID string) IsuConditionPoster {
	ctx, cancel := context.WithCancel(context.Background())
	return IsuConditionPoster{*targetURL, isuUUID, ctx, cancel}
}

func (m *IsuConditionPoster) KeepPosting() {
	targetURL := m.TargetURL
	targetURL.Path = path.Join(targetURL.Path, "/api/condition/", m.IsuUUID)
	randEngine := rand.New(rand.NewSource(0))

	nowTime := time.Now()
	randEngine.Seed(nowTime.UnixNano()/1000000000 + 961054102) // 乱数初期化（逆算できるように）

	timer := time.NewTicker(postingIntervalSec * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-timer.C:
		}

		conditions := []IsuConditionRequest{}
		for i := 0; i < postingIntervalSec; i++ {
			cond := IsuConditionRequest{
				IsSitting: (randEngine.Intn(100) <= 70),
				Condition: fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v",
					(randEngine.Intn(2) == 0),
					(randEngine.Intn(2) == 0),
					(randEngine.Intn(2) == 0),
				),
				Message:   "テストメッセージです",
				Timestamp: nowTime.Unix(),
			}
			conditions = append(conditions, cond)

			nowTime = nowTime.Add(1 * time.Second)
		}

		conditionsJSON, err := json.Marshal(conditions)
		if err != nil {
			log.Error(err)
			continue
		}

		func() {
			httpReq, err := http.NewRequest(
				http.MethodPost, targetURL.String(),
				bytes.NewBuffer(conditionsJSON),
			)
			if err != nil {
				log.Error(err)
				return // goto next loop
			}
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("User-Agent", "JIA-Members-Client-MOCK/1.0")
			resp, err := http.DefaultClient.Do(httpReq)
			if err != nil {
				log.Error(err)
				return // goto next loop
			}
			defer resp.Body.Close()

			if resp.StatusCode != 202 {
				log.Errorf("`POST %s` returned unexpected status code `%s`", targetURL.String(), resp.Status)
				return // goto next loop
			}
		}()
	}
}

func (m *IsuConditionPoster) StopPosting() {
	m.cancelFunc()
}
