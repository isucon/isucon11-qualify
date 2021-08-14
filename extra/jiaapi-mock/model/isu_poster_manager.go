package model

import (
	"sync"
)

type IsuConditionPosterManager struct {
	activatedIsu    map[string]IsuConditionPoster
	activatedIsuMtx sync.Mutex
}

func NewIsuConditionPosterManager() *IsuConditionPosterManager {
	activatedIsu := make(map[string]IsuConditionPoster)
	return &IsuConditionPosterManager{activatedIsu, sync.Mutex{}}
}

func (m *IsuConditionPosterManager) StartPosting(targetURL string, isuUUID string) error {
	key := getKey(targetURL, isuUUID)

	conflict := func() bool {
		m.activatedIsuMtx.Lock()
		defer m.activatedIsuMtx.Unlock()
		if _, ok := m.activatedIsu[key]; ok {
			return true
		}
		m.activatedIsu[key] = NewIsuConditionPoster(targetURL, isuUUID)
		return false
	}()
	if !conflict {
		isu := m.activatedIsu[key]
		go isu.KeepPosting()
	}
	return nil
}
