package model

import (
	"context"
	"sync"
)

//一つのIsuにつき、一つの送信用スレッドがある
type Isu struct {
	Mutex          sync.Mutex
	JIAIsuUUID     string `json:"jia_isu_uuid"`
	Name           string `json:"name"`
	ImageNAme      string `json:"-"`
	JIACatalogID   string `json:"jia_catalog_id"`
	Character      string `json:"character"`
	isDeleted      bool
	deactivateFunc context.CancelFunc
	//TODO: その他postスレッドと通信するためのchannel
	Owner      *User
	Conditions []IsuCondition
}

func (isu *Isu) IsDeleted() bool {
	return isu.isDeleted
}

func (isu *Isu) IsDeleted() bool {
	return isu.isDeleted
}

func (isu *Isu) Delete() {
	if !isu.isDeleted {
		isu.isDeleted = false
		isu.deactivateFunc()
	}
}
