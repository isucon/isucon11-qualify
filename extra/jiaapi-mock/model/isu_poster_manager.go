package model

import (
	"net/url"
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

func (m *IsuConditionPosterManager) StartPosting(targetURL *url.URL, isuUUID string) error {
	conflict := func() bool {
		m.activatedIsuMtx.Lock()
		defer m.activatedIsuMtx.Unlock()
		if _, ok := m.activatedIsu[isuUUID]; ok {
			return true
		}
		m.activatedIsu[isuUUID] = NewIsuConditionPoster(targetURL, isuUUID)
		return false
	}()
	if !conflict {
		isu := m.activatedIsu[isuUUID]
		go isu.KeepPosting()
	}
	return nil
}
