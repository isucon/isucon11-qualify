package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreStartBenchmark        score.ScoreTag = "StartBenchmark       "
	ScoreNormalUserInitialize  score.ScoreTag = "NormalUserInitialize "
	ScoreGraphExcellent        score.ScoreTag = "GraphExcellent       "
	ScoreGraphGood             score.ScoreTag = "GraphGood            "
	ScoreGraphNormal           score.ScoreTag = "GraphNormal          "
	ScoreGraphBad              score.ScoreTag = "GraphBad             "
	ScoreGraphWorst            score.ScoreTag = "GraphWorst           "
	ScoreReadInfoCondition     score.ScoreTag = "ReadInfoCondition    "
	ScoreReadWarningCondition  score.ScoreTag = "ReadWarningCondition "
	ScoreReadCriticalCondition score.ScoreTag = "ReadCriticalCondition"
)
