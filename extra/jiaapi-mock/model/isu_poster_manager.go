package model

import (
	"net/url"
	"sync"
)

type IsuConditionPosterManager struct {
	activatedIsu      map[string]IsuConditionPoster
	activatedIsuUUIDs []string
	activatedIsuMtx   sync.Mutex
}

func NewIsuConditionPosterManager() *IsuConditionPosterManager {
	activatedIsu := make(map[string]IsuConditionPoster)
	return &IsuConditionPosterManager{activatedIsu, []string{}, sync.Mutex{}}
}

func (m *IsuConditionPosterManager) StartPosting(targetURL *url.URL, isuUUID string) error {
	key := getKey(targetURL.String(), isuUUID)

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
		m.activatedIsuUUIDs = append(m.activatedIsuUUIDs, isuUUID)
		go isu.KeepPosting()
	}
	return nil
}

func (m *IsuConditionPosterManager) IsActivatedUUID(isuUUID string) bool {
	for _, activatedUUID := range m.activatedIsuUUIDs {
		if isuUUID == activatedUUID {
			return true
		}
	}
	return false
}
