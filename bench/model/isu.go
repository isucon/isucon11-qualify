package model

//enum
type IsuStateChange int

//posterスレッドとシナリオスレッドとの通信に必要な情報
//複数回posterスレッドが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type IsuPosterChan struct {
	JIAIsuUUID    string
	activateChan  chan bool           //posterスレッド -> シナリオスレッド
	isuChan       chan IsuStateChange //シナリオスレッド->posterスレッド
	conditionChan chan IsuCondition   //posterスレッド -> シナリオスレッド
}

//一つのIsuにつき、一つの送信用スレッドがある
//IsuはISU協会スレッドからも読み込まれる
type Isu struct {
	Owner             *User
	JIAIsuUUID        string
	Name              string
	ImageName         string
	JIACatalogID      string
	Character         string
	isWantDeactivated bool           //シナリオ上でDeleteリクエストを送ったかどうか
	isDeactivated     bool           //実際にdeactivateされているか
	PosterChan        *IsuPosterChan //ISU協会はこれを使ってposterスレッドを起動、posterスレッドはこれを使って通信
	Conditions        []IsuCondition //シナリオスレッドからのみ参照
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
	case v, ok := <-isu.PosterChan.activateChan:
		isu.isDeactivated = !ok || !v //Isu協会スレッドの終了 || deactivateされた
	default:
	}
	return isu.isDeactivated
}

//シナリオスレッドからのみ参照
func (isu *Isu) IsWantDeactivated() bool {
	return isu.isWantDeactivated
}

//シナリオスレッドからのみ参照
func (isu *Isu) WantDeactivated() {
	isu.isWantDeactivated = true
}
