package model

//enum
type IsuStateChange int

//postingスレッドとシナリオスレッドとの通信に必要な情報
//複数回postingスレッドが起動するかもしれないのでcloseしない
//当然リソースリークするがベンチマーカーは毎回落とすので問題ない
type IsuPostingChan struct {
	JIAIsuUUID    string
	activateChan  chan bool           //postingスレッド -> シナリオスレッド
	isuChan       chan IsuStateChange //シナリオスレッド->postingスレッド
	conditionChan chan IsuCondition   //postingスレッド -> シナリオスレッド
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
	postingChan   *IsuPostingChan //ISU協会はこれを使ってpostingスレッドを起動、postingスレッドはこれを使って通信
	//deactivateFunc context.CancelFunc //Isu協会activate/deactivateスレッドからのみ参照 //TODO: Isu協会側でデータを持つ
	Conditions []IsuCondition //シナリオスレッドからのみ参照
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
	case v, ok := <-isu.postingChan.activateChan:
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
