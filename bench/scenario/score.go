package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreStartBenchmark        score.ScoreTag = "0.StartBenchmark       "
	ScoreGraphExcellent        score.ScoreTag = "1.GraphExcellent       "
	ScoreGraphGood             score.ScoreTag = "2.GraphGood            "
	ScoreGraphNormal           score.ScoreTag = "3.GraphNormal          "
	ScoreGraphBad              score.ScoreTag = "4.GraphBad             "
	ScoreGraphWorst            score.ScoreTag = "5.GraphWorst           "
	ScoreReadInfoCondition     score.ScoreTag = "6.ReadInfoCondition    "
	ScoreReadWarningCondition  score.ScoreTag = "7.ReadWarningCondition "
	ScoreReadCriticalCondition score.ScoreTag = "8.ReadCriticalCondition"
	ScoreNormalUserInitialize  score.ScoreTag = "_.NormalUserInitialize " //scoreが0のもの
	ScoreViewerInitialize      score.ScoreTag = "_.ViewerInitialize     "
	ScoreViewerDropout         score.ScoreTag = "_.ViewerDropout        "
)
