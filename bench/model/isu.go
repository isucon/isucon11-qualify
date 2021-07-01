package model

import "fmt"

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
	Owner         *User
	JIAIsuUUID    string `json:"jia_isu_uuid"`
	Name          string `json:"name"`
	ImageName     string `json:"-"`
	JIACatalogID  string `json:"jia_catalog_id"`
	Character     string `json:"character"`
	isWantDeleted bool   //シナリオスレッドからのみ参照
	isDeactivated bool
	PosterChan    *IsuPosterChan //ISU協会はこれを使ってposterスレッドを起動、posterスレッドはこれを使って通信
	Conditions    []IsuCondition //シナリオスレッドからのみ参照
}

//新しいISUの生成
//senarioのNewIsu以外からは呼び出さないこと！
//戻り値を使ってbackendにpostする必要あり
//戻り値をISU協会にIsu*を登録する必要あり
//戻り値をownerに追加する必要あり
func NewRandomIsuRaw(owner *User) *Isu {
	id := fmt.Sprintf("randomid-%s-%d", owner.UserID, len(owner.IsuListOrderByCreatedAt))     //TODO: ちゃんと生成する
	name := fmt.Sprintf("randomname-%s-%d", owner.UserID, len(owner.IsuListOrderByCreatedAt)) //TODO: ちゃんと生成する
	isu := &Isu{
		Owner:         owner,
		JIAIsuUUID:    id,
		Name:          name,
		ImageName:     "dafault-image", //TODO: ちゃんとデータに合わせる
		JIACatalogID:  "",              //TODO:
		Character:     "",              //TODO:
		isWantDeleted: false,
		isDeactivated: true,
		PosterChan: &IsuPosterChan{
			JIAIsuUUID:    id,
			activateChan:  make(chan bool),
			isuChan:       make(chan IsuStateChange, 1),
			conditionChan: make(chan IsuCondition),
		},
		Conditions: []IsuCondition{},
	}

	return isu
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
func (isu *Isu) IsDeleted() bool {
	return isu.isWantDeleted
}

//シナリオスレッドからのみ参照
func (isu *Isu) WantDelete() {
	isu.isWantDeleted = true
}
