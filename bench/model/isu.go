package model

import "fmt"

//enum
type IsuStateChange int

const (
	IsuStateChangeNone IsuStateChange = iota
	IsuStateChangeClear
	IsuStateChangeDetectOverweight
	IsuStateChangeClearAndDetect
	IsuStateChangeBad
	IsuStateChangeDelete //椅子を削除する
)

//posterスレッドとシナリオスレッドとの通信に必要な情報
//ISU協会はこれを使ってposterスレッドを起動、posterスレッドはこれを使って通信
//複数回posterスレッドが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type StreamsForPoster struct {
	ActiveChan    chan<- bool
	StateChan     <-chan IsuStateChange
	ConditionChan chan<- IsuCondition
}

//posterスレッドとシナリオスレッドとの通信に必要な情報
//複数回posterスレッドが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type StreamsForScenario struct {
	activeChan    <-chan bool
	StateChan     chan<- IsuStateChange
	ConditionChan <-chan IsuCondition
}

//一つのIsuにつき、一つの送信用スレッドがある
//IsuはISU協会スレッドからも読み込まれる
type Isu struct {
	Owner              *User
	JIAIsuUUID         string
	Name               string
	ImageName          string
	JIACatalogID       string
	Character          string
	IsWantDeactivated  bool                //シナリオ上でDeleteリクエストを送ったかどうか
	isDeactivated      bool                //実際にdeactivateされているか
	StreamsForScenario *StreamsForScenario //posterスレッドとの通信
	Conditions         []IsuCondition      //シナリオスレッドからのみ参照
}

//新しいISUの生成
//senarioのNewIsu以外からは呼び出さないこと！
//戻り値を使ってbackendにpostする必要あり
//戻り値をISU協会にIsu*を登録する必要あり
//戻り値をownerに追加する必要あり
func NewRandomIsuRaw(owner *User) (*Isu, *StreamsForPoster, error) {
	activeChan := make(chan bool)
	stateChan := make(chan IsuStateChange, 1)
	conditionChan := make(chan IsuCondition)

	id := fmt.Sprintf("randomid-%s-%d", owner.UserID, len(owner.IsuListOrderByCreatedAt))     //TODO: ちゃんと生成する
	name := fmt.Sprintf("randomname-%s-%d", owner.UserID, len(owner.IsuListOrderByCreatedAt)) //TODO: ちゃんと生成する
	isu := &Isu{
		Owner:             owner,
		JIAIsuUUID:        id,
		Name:              name,
		ImageName:         "dafault-image", //TODO: ちゃんとデータに合わせる
		JIACatalogID:      "",              //TODO:
		Character:         "",              //TODO:
		IsWantDeactivated: false,
		isDeactivated:     true,
		StreamsForScenario: &StreamsForScenario{
			activeChan:    activeChan,
			StateChan:     stateChan,
			ConditionChan: conditionChan,
		},
		Conditions: []IsuCondition{},
	}

	streamsForPoster := &StreamsForPoster{
		ActiveChan:    activeChan,
		StateChan:     stateChan,
		ConditionChan: conditionChan,
	}
	return isu, streamsForPoster, nil
}

//シナリオスレッドからのみ参照
func (isu *Isu) IsDeactivated() bool {
	select {
	case v, ok := <-isu.StreamsForScenario.activeChan:
		isu.isDeactivated = !ok || !v //Isu協会スレッドの終了 || deactivateされた
	default:
	}
	return isu.isDeactivated
}
