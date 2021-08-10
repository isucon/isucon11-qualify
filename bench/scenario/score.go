package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreStartBenchmark        score.ScoreTag = "0.StartBenchmark       "
	ScoreNormalUserInitialize  score.ScoreTag = "1.NormalUserInitialize "
	ScoreGraphExcellent        score.ScoreTag = "2.GraphExcellent       "
	ScoreGraphGood             score.ScoreTag = "3.GraphGood            "
	ScoreGraphNormal           score.ScoreTag = "4.GraphNormal          "
	ScoreGraphBad              score.ScoreTag = "5.GraphBad             "
	ScoreGraphWorst            score.ScoreTag = "6.GraphWorst           "
	ScoreReadInfoCondition     score.ScoreTag = "7.ReadInfoCondition    "
	ScoreReadWarningCondition  score.ScoreTag = "8.ReadWarningCondition "
	ScoreReadCriticalCondition score.ScoreTag = "9.ReadCriticalCondition"
)
