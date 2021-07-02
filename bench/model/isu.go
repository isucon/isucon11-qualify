package model

//enum
type IsuStateChange int

//posterスレッドとシナリオスレッドとの通信に必要な情報
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
	StreamsForScenario *StreamsForScenario //ISU協会はこれを使ってposterスレッドを起動、posterスレッドはこれを使って通信
	Conditions         []IsuCondition      //シナリオスレッドからのみ参照
}

func NewIsu() *Isu {
	v := &Isu{
		isDeactivated: true,
		//TODO: ポインタやchanの初期化
	}
	//TODO: ISU協会にIsu*を登録

	return v
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
