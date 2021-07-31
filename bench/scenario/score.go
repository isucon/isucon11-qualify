package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreNormalUserInitialize  score.ScoreTag = "NormalUserInitialize      "
	ScoreNormalUserLoop        score.ScoreTag = "NormalUserLoop            "
	ScoreCompanyUserInitialize score.ScoreTag = "CompanyUserInitialize     "
	ScoreCompanyUserLoop       score.ScoreTag = "CompanyUserLoop           "
	ScoreGraphExcellent        score.ScoreTag = "ScoreGraphExcellent       "
	ScoreGraphGood             score.ScoreTag = "ScoreGraphGood            "
	ScoreGraphNormal           score.ScoreTag = "ScoreGraphNormal          "
	ScoreGraphBad              score.ScoreTag = "ScoreGraphBad             "
	ScoreGraphWorst            score.ScoreTag = "ScoreGraphWorst           "
	ScoreReadInfoCondition     score.ScoreTag = "ScoreReadInfoCondition    "
	ScoreReadWarningCondition  score.ScoreTag = "ScoreReadWarningCondition "
	ScoreReadCriticalCondition score.ScoreTag = "ScoreReadCriticalCondition"
)
