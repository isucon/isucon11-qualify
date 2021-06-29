package model

import (
	"context"
	"sync"
)

type IsuConditionArray struct {
	Mutex      sync.Mutex
	Conditions []IsuCondition //appendでアドレスが変わりうるので、mutexを取っている間にコピーすること
}

//一つのIsuにつき、一つの送信用スレッドがある
type Isu struct {
	Owner          *User
	JIAIsuUUID     string             `json:"jia_isu_uuid"`
	Name           string             `json:"name"`
	ImageName      string             `json:"-"`
	JIACatalogID   string             `json:"jia_catalog_id"`
	Character      string             `json:"character"`
	isWantDeleted  bool               //シナリオスレッドからのみ参照
	deactivateFunc context.CancelFunc //Isu協会activate/deactivateスレッドからのみ参照
	//TODO: その他postスレッドと通信するためのchannel //シナリオスレッド->postingスレッド
	Conditions *IsuConditionArray //シナリオスレッドからread、postingスレッドからwrite
}

//Isu協会activate/deactivateスレッドからのみ呼び出される
func (isu *Isu) Deactivate() {
	if isu.deactivateFunc != nil {
		isu.deactivateFunc()
		isu.deactivateFunc = nil
	}
}

//シナリオスレッドからのみ参照
func (isu *Isu) IsDeleted() bool {
	return isu.isWantDeleted
}

//シナリオスレッドからのみ参照
func (isu *Isu) WantDelete() {
	isu.isWantDeleted = true
}
