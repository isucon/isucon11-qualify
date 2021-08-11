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
	// key: isuID, value: timestamp
	verifiedConditionsInTrend      map[int]int64
	verifiedConditionsInTrendMutex sync.RWMutex
}

func NewViewer(agent *agent.Agent) Viewer {
	return Viewer{
		ErrorCount:                0,
		ViewedUpdatedCount:        0,
		Agent:                     agent,
		verifiedConditionsInTrend: make(map[int]int64, 8192),
	}
}

func (v *Viewer) SetVerifiedCondition(id int, timestamp int64) {
	v.verifiedConditionsInTrendMutex.Lock()
	defer v.verifiedConditionsInTrendMutex.Unlock()
	v.verifiedConditionsInTrend[id] = timestamp
}

func (v *Viewer) ConditionAlreadyVerified(id int, timestamp int64) bool {
	v.verifiedConditionsInTrendMutex.RLock()
	defer v.verifiedConditionsInTrendMutex.RUnlock()
	t, exist := v.verifiedConditionsInTrend[id]
	if exist && t == timestamp {
		return true
	}
	return false
}

func (v *Viewer) ConditionIsUpdated(id int, timestamp int64) bool {
	v.verifiedConditionsInTrendMutex.RLock()
	defer v.verifiedConditionsInTrendMutex.RUnlock()
	t := v.verifiedConditionsInTrend[id]
	return t < timestamp
}

func (v *Viewer) NumOfIsu() int {
	v.verifiedConditionsInTrendMutex.RLock()
	defer v.verifiedConditionsInTrendMutex.RUnlock()
	return len(v.verifiedConditionsInTrend)
}
