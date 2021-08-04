package model

import (
	"sync"

	"github.com/isucon/isucandar/agent"
)

//基本的には一つのシナリオ Goroutine が一つの Viewer を占有する
type Viewer struct {
	ErrorCount         uint
	ViewedUpdatedCount uint
	Agent              *agent.Agent

	// GET trend にて既に確認したconditionを格納するのに利用
	verifiedConditionsInTrend      map[isuIDAndConditionTimestamp]struct{}
	verifiedConditionsInTrendMutex sync.RWMutex
}

type isuIDAndConditionTimestamp struct {
	IsuID              int
	ConditionTimestamp int64
}

func NewViewer(agent *agent.Agent) Viewer {
	return Viewer{
		ErrorCount:                0,
		ViewedUpdatedCount:        0,
		Agent:                     agent,
		verifiedConditionsInTrend: make(map[isuIDAndConditionTimestamp]struct{}, 8192),
	}
}

func (v *Viewer) SetVerifiedCondition(id int, timestamp int64) {
	v.verifiedConditionsInTrendMutex.Lock()
	defer v.verifiedConditionsInTrendMutex.Unlock()
	v.verifiedConditionsInTrend[isuIDAndConditionTimestamp{id, timestamp}] = struct{}{}
}

func (v *Viewer) ConditionAlreadyVerified(id int, timestamp int64) bool {
	v.verifiedConditionsInTrendMutex.RLock()
	defer v.verifiedConditionsInTrendMutex.RUnlock()
	_, exist := v.verifiedConditionsInTrend[isuIDAndConditionTimestamp{id, timestamp}]
	return exist
}
