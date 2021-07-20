package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreNormalUserInitialize  score.ScoreTag = "nu-i"
	ScoreNormalUserLoop        score.ScoreTag = "nu-l"
	ScoreCompanyUserInitialize score.ScoreTag = "cu-i"
	ScoreCompanyUserLoop       score.ScoreTag = "cu-l"
	ScorePostConditionInfo     score.ScoreTag = "pc-i"
	ScorePostConditionWarning  score.ScoreTag = "pc-w"
	ScorePostConditionCritical score.ScoreTag = "pc-c"
)
