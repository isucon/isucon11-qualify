package model

import (
	"sync"
)

type IsuStateChange int

type IsuConditionArray struct {
	Mutex      sync.Mutex
	Conditions []IsuCondition //appendでアドレスが変わりうるので、mutexを取っている間にコピーすること
}

//一つのIsuにつき、一つの送信用スレッドがある
//IsuはISU協会スレッドからも読み込まれる
type Isu struct {
	Owner         *User
	JIAIsuUUID    string `json:"jia_isu_uuid"`
	Name          string `json:"name"`
	ImageName     string `json:"-"`
	JIACatalogID  string `json:"jia_catalog_id"`
	Character     string `json:"character"`
	isWantDeleted bool   //シナリオスレッドからのみ参照
	isDeactivated bool
	activateChan  chan bool //Isu協会 -> シナリオスレッド
	//deactivateFunc context.CancelFunc //Isu協会activate/deactivateスレッドからのみ参照 //TODO: Isu協会側でデータを持つ
	isuChan    chan IsuStateChange //シナリオスレッド->postingスレッド
	Conditions *IsuConditionArray  //シナリオスレッドからread、postingスレッドからwrite
}

func NewIsu() *Isu {
	v := &Isu{
		isDeactivated: true,
		//TODO: ポインタやchanの初期化
	}
	//TODO: ISU協会にIsu*を登録

	return v
}

func (isu *Isu) IsDeactivated() bool {
	select {
	case v, ok := <-isu.activateChan:
		isu.isDeactivated = !ok || !v //Isu協会スレッドの終了 || deactivateされた
	default:
	}
	return isu.isDeactivated
}

//シナリオスレッドからのみ参照
func (isu *Isu) IsDeleted() bool {
	return isu.isWantDeleted
}

//シナリオスレッドからのみ参照
func (isu *Isu) WantDelete() {
	isu.isWantDeleted = true
}
