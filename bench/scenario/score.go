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

func SetScoreTags(scoreTable score.ScoreTable) {
	setScoreTag(scoreTable, ScoreStartBenchmark)
	setScoreTag(scoreTable, ScoreGraphExcellent)
	setScoreTag(scoreTable, ScoreGraphGood)
	setScoreTag(scoreTable, ScoreGraphNormal)
	setScoreTag(scoreTable, ScoreGraphBad)
	setScoreTag(scoreTable, ScoreGraphWorst)
	setScoreTag(scoreTable, ScoreTodayGraphExcellent)
	setScoreTag(scoreTable, ScoreTodayGraphGood)
	setScoreTag(scoreTable, ScoreTodayGraphNormal)
	setScoreTag(scoreTable, ScoreTodayGraphBad)
	setScoreTag(scoreTable, ScoreTodayGraphWorst)
	setScoreTag(scoreTable, ScoreReadInfoCondition)
	setScoreTag(scoreTable, ScoreReadWarningCondition)
	setScoreTag(scoreTable, ScoreReadCriticalCondition)
	setScoreTag(scoreTable, ScoreIsuInitialize)
	setScoreTag(scoreTable, ScoreNormalUserInitialize)
	setScoreTag(scoreTable, ScoreViewerInitialize)
	setScoreTag(scoreTable, ScoreViewerDropout)
	setScoreTag(scoreTable, ScoreRepairIsu)
	setScoreTag(scoreTable, ScorePostInfoCondition)
	setScoreTag(scoreTable, ScorePostWarningCondition)
	setScoreTag(scoreTable, ScorePostCriticalCondition)
}

func setScoreTag(scoreTable score.ScoreTable, tag score.ScoreTag) {
	if _, ok := scoreTable[tag]; !ok {
		scoreTable[tag] = 0
	}
}
