package model

import (
	"context"

	"github.com/google/uuid"
	"github.com/isucon/isucon11-qualify/extra/initial-data/random"
)

//enum
type IsuStateChange int

const (
	IsuStateChangeNone IsuStateChange = iota
	IsuStateChangeBad
	IsuStateChangeDelete           //椅子を削除する
	IsuStateChangeClear            = 1 << 3
	IsuStateChangeDetectOverweight = 1 << 4
	IsuStateChangeRepair           = 1 << 5
)

//posterスレッドとシナリオスレッドとの通信に必要な情報
//ISU協会はこれを使ってposterスレッドを起動、posterスレッドはこれを使って通信
//複数回posterスレッドが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type StreamsForPoster struct {
	ActiveChan    chan<- bool
	StateChan     <-chan IsuStateChange
	ConditionChan chan<- []IsuCondition
}

//posterスレッドとシナリオスレッドとの通信に必要な情報
//複数回posterスレッドが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type StreamsForScenario struct {
	activeChan    <-chan bool
	StateChan     chan<- IsuStateChange
	ConditionChan <-chan []IsuCondition
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
	Conditions         IsuConditionArray   //シナリオスレッドからのみ参照
}

//新しいISUの生成
//scenarioのNewIsu以外からは呼び出さないこと！
//戻り値を使ってbackendにpostする必要あり
//戻り値をISU協会にIsu*を登録する必要あり
//戻り値をownerに追加する必要あり
func NewRandomIsuRaw(owner *User) (*Isu, *StreamsForPoster, error) {
	activeChan := make(chan bool, 1) //容量1以上ないとposterがブロックするので、必ず1以上
	stateChan := make(chan IsuStateChange, 1)
	conditionChan := make(chan []IsuCondition, 10)

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, nil, err
	}
	isu := &Isu{
		Owner:             owner,
		JIAIsuUUID:        id.String(),
		Name:              random.IsuName(),
		ImageName:         "NoImage.png",                          //TODO: ちゃんとデータに合わせる
		JIACatalogID:      "550e8400-e29b-41d4-a716-446655440000", //TODO:
		Character:         random.Character(),
		IsWantDeactivated: false,
		isDeactivated:     true,
		StreamsForScenario: &StreamsForScenario{
			activeChan:    activeChan,
			StateChan:     stateChan,
			ConditionChan: conditionChan,
		},
		Conditions: NewIsuConditionArray(),
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

func (isu *Isu) getConditionFromChan(ctx context.Context, userConditionBuffer *IsuConditionTreeSet) {
	for {
		select {
		case <-ctx.Done():
			return
		case conditions, ok := <-isu.StreamsForScenario.ConditionChan:
			if !ok {
				return
			}
			for i := range conditions {
				isu.Conditions.Add(&conditions[i]) //copyなので問題ない
			}
			if userConditionBuffer != nil {
				for i := range conditions {
					userConditionBuffer.Add(&conditions[i])
				}
			}
		default:
			return
		}
	}
}
