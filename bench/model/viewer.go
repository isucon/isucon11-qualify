package model

import (
	"net/http"
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
	verifiedConditionsInTrend map[int]int64

	// asset名がキー、そのhashが値
	staticCacheMx    sync.Mutex
	StaticCachedHash map[string]uint32
}

func NewViewer(agent *agent.Agent) Viewer {
	return Viewer{
		ErrorCount:                0,
		ViewedUpdatedCount:        0,
		Agent:                     agent,
		verifiedConditionsInTrend: make(map[int]int64, 700),
		staticCacheMx:             sync.Mutex{},
		StaticCachedHash:          make(map[string]uint32),
	}
}

func (v *Viewer) SetVerifiedCondition(id int, timestamp int64) {
	v.verifiedConditionsInTrend[id] = timestamp
}

func (v *Viewer) ConditionAlreadyVerified(id int, timestamp int64) bool {
	t, exist := v.verifiedConditionsInTrend[id]
	if exist && t == timestamp {
		return true
	}
	return false
}

func (v *Viewer) ConditionIsUpdated(id int, timestamp int64) bool {
	t := v.verifiedConditionsInTrend[id]
	return t < timestamp
}

func (v *Viewer) NumOfIsu() int {
	return len(v.verifiedConditionsInTrend)
}

func (v *Viewer) GetAgent() *agent.Agent {
	return v.Agent
}

func (v *Viewer) SetStaticCache(path string, hash uint32) {
	v.staticCacheMx.Lock()
	defer v.staticCacheMx.Unlock()

	if v.StaticCachedHash == nil {
		v.StaticCachedHash = map[string]uint32{}
	}

	v.StaticCachedHash[path] = hash
}

func (v *Viewer) GetStaticCache(path string, _ *http.Request) (uint32, bool) {
	v.staticCacheMx.Lock()
	defer v.staticCacheMx.Unlock()

	if v.StaticCachedHash == nil {
		v.StaticCachedHash = map[string]uint32{}
	}

	hash, exist := v.StaticCachedHash[path]
	return hash, exist
}

func (v *Viewer) ClearStaticCache() {
	v.staticCacheMx.Lock()
	defer v.staticCacheMx.Unlock()


	v.StaticCachedHash = map[string]uint32{}
}
