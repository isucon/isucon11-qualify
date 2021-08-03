package scenario

// score関係の定数には「Score」プレフィックスをつける

type IScoreGraphTimestampCount struct {
	Excellent int
	Good      int
	Normal    int
	Bad       int
	Worst     int
}

// 現状 virtualTimeMulti は 30000、timeout は 5ms、PostContentNum は 100 なので、仮想時間において 150s につき 100 件のデータが送られる。
// 一時間は 60 * 60 秒なので、 60 * 60 * (100 / 150) = 24 * 100 個のデータが入りうる(timeoutよりはやくレスポンスが帰ってきた時はこれより多いがとりあえず気にしない)。
var ScoreGraphTimestampCount = IScoreGraphTimestampCount{
	Excellent: 2000,
	Good:      1500,
	Normal:    1000,
	Bad:       500,
	Worst:     0,
}

// 一日の秒数
const OneDay = 24 * 60 * 60

// GET /api/isu/:id/condition の
const ConditionPagingStep = 10

const SignoutPercentage = 20
