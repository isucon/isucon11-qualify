package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreNormalUserInitialize  score.ScoreTag = "NormalUserInitialize "
	ScoreNormalUserLoop        score.ScoreTag = "NormalUserLoop       "
	ScoreCompanyUserInitialize score.ScoreTag = "CompanyUserInitialize"
	ScoreCompanyUserLoop       score.ScoreTag = "CompanyUserLoop      "
	ScorePostConditionInfo     score.ScoreTag = "PostConditionInfo    "
	ScorePostConditionWarning  score.ScoreTag = "PostConditionWarning "
	ScorePostConditionCritical score.ScoreTag = "PostConditionCritical"
)
