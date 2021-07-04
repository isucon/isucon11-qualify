package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

var (
	ScoreAuth                  score.ScoreTag = "auth"
	ScorePostConditionInfo     score.ScoreTag = "pc-i"
	ScorePostConditionWarning  score.ScoreTag = "pc-w"
	ScorePostConditionCritical score.ScoreTag = "pc-c"
)
