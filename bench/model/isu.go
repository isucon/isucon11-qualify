package model

import (
	"context"
	"crypto/md5"
	"github.com/isucon/isucon11-qualify/bench/service"
	"io/ioutil"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/isucon/isucon11-qualify/bench/random"
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

//poster Goroutineとシナリオ Goroutineとの通信に必要な情報
//ISU協会はこれを使ってposter Goroutineを起動、poster Goroutineはこれを使って通信
//複数回poster Goroutineが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type StreamsForPoster struct {
	ActiveChan    chan<- bool
	StateChan     <-chan IsuStateChange
	ConditionChan chan<- []IsuCondition
}

//poster Goroutineとシナリオ Goroutineとの通信に必要な情報
//複数回poster Goroutineが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type StreamsForScenario struct {
	activeChan    <-chan bool
	StateChan     chan<- IsuStateChange
	ConditionChan <-chan []IsuCondition
}

//一つのIsuにつき、一つの送信用 Goroutineがある
//IsuはISU協会 Goroutineからも読み込まれる
type Isu struct {
	Owner              *User               `json:"-"`
	JIAIsuUUID         string              `json:"jia_isu_uuid"`
	Name               string              `json:"name"`
	ImageHash          [md5.Size]byte      `json:"image_file_hash"`
	JIACatalogID       string              `json:"-"`
	Character          string              `json:"character"`
	IsWantDeactivated  bool                `json:"-"`          //シナリオ上でDeleteリクエストを送ったかどうか
	isDeactivated      bool                `json:"-"`          //実際にdeactivateされているか
	StreamsForScenario *StreamsForScenario `json:"-"`          //poster Goroutineとの通信
	Conditions         IsuConditionArray   `json:"conditions"` //シナリオ Goroutineからのみ参照
	CreatedAt          time.Time           `json:"created_at"`
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
		ImageHash:         defaultIconHash,
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

//シナリオ Goroutineからのみ参照
func (isu *Isu) IsDeactivated() bool {
	select {
	case v, ok := <-isu.StreamsForScenario.activeChan:
		isu.isDeactivated = !ok || !v //Isu協会 Goroutineの終了 || deactivateされた
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

var defaultIconHash [md5.Size]byte

const defaultIconFilePath = "./images/default.jpg"

func init() {
	image, err := ioutil.ReadFile(defaultIconFilePath)
	if err != nil {
		log.Fatalf("failed to read default icon: %v", err)
	}
	defaultIconHash = md5.Sum(image)
}

func (isu *Isu) ToService() *service.Isu {
	return &service.Isu{
		JIAIsuUUID: isu.JIAIsuUUID,
		Name:       isu.Name,
		Character:  isu.Character,
	}
}
