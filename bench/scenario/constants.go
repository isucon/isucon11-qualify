package scenario

// score関係の定数には「Score」プレフィックスをつける

type IScoreGraphTimestampCount struct {
	Good   int
	Normal int
	Bad    int
	Worst  int
}

// 現状 virtualTimeMulti は 30000、で timeout は 5ms、より timeout の間隔で仮想時間では 1500s たっている。
// PostConditionIntervalSecond が 60s なので timeout の時間に最高で 1500s / 60s = 25 個の condition が存在する
// しかし PostConditionNum が 10 なので backend がめちゃくちゃ早くレスポンスを返さないと、理論上存在する 25個の condition は 10個になり点数のソースを失う
// ScoreGraph で関心がある1時間の condition の timestamp の数については、理論値(60sに一件)は max60個、timeout はするが全件 insert するのは 24個
var ScoreGraphTimestampCount = IScoreGraphTimestampCount{
	Good:   20,
	Normal: 10,
	Bad:    5,
	Worst:  0,
}

// 一日の秒数
const OneDay = 24 * 60 * 60

// GET /api/isu/:id/condition の
const ConditionPagingStep = 1

const SignoutPercentage = 20

// ReadCondition/PostCondition 系のスコアタグが何件ごとに付与されるか
const ReadConditionTagStep = 50

// User を増やすかどうかの閾値
const AddUserStep = 500

// User を増やすとき何人増やすか
const AddUserCount = 1

// Viewer が何回以上エラーに遭遇したら drop するか
const ViewerDropCount = 1

type PageType int

const (
	HomePage PageType = iota
	IsuDetailPage
	IsuConditionPage
	IsuGraphPage
	RegisterPage
	TrendPage
)

// ユーザーがもってるISUの数の上限
const IsuCountMax = 9

// 1ユーザーのループが何回回れば Viewer が増えるか
const ViewerAddLoopStep = 1

// Viewer のユーザー数に対する上限
const ViewerLimitPerUser = 3

// ユーザーが追加されるとき、発生していて良い Timeout のユーザー数に対する上限
const TimeoutLimitPerUser = 20

// GET /api/isu/:id/graph と GET /api/condition/:id の間で許される condition 反映の遅延
const ConditionDelayTime = 1
