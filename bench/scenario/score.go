package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreNormalUserInitialize  score.ScoreTag = "nu-i"
	ScoreNormalUserLoop        score.ScoreTag = "nu-l"
	ScorePostConditionInfo     score.ScoreTag = "pc-i"
	ScorePostConditionWarning  score.ScoreTag = "pc-w"
	ScorePostConditionCritical score.ScoreTag = "pc-c"
)
