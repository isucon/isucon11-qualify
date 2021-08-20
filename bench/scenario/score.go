package scenario

import "github.com/isucon/isucandar/score"

// スコアタグの管理

const (
	ScoreStartBenchmark        score.ScoreTag = "00.StartBenchmark       "
	ScoreGraphGood             score.ScoreTag = "01.GraphGood            "
	ScoreGraphNormal           score.ScoreTag = "02.GraphNormal          "
	ScoreGraphBad              score.ScoreTag = "03.GraphBad             "
	ScoreGraphWorst            score.ScoreTag = "04.GraphWorst           "
	ScoreTodayGraphGood        score.ScoreTag = "05.TodayGraphGood       "
	ScoreTodayGraphNormal      score.ScoreTag = "06.TodayGraphNormal     "
	ScoreTodayGraphBad         score.ScoreTag = "07.TodayGraphBad        "
	ScoreTodayGraphWorst       score.ScoreTag = "08.TodayGraphWorst      "
	ScoreReadInfoCondition     score.ScoreTag = "09.ReadInfoCondition    "
	ScoreReadWarningCondition  score.ScoreTag = "10.ReadWarningCondition "
	ScoreReadCriticalCondition score.ScoreTag = "11.ReadCriticalCondition"
	ScoreIsuInitialize         score.ScoreTag = "_1.IsuInitialize        " //scoreが0のもの
	ScoreNormalUserInitialize  score.ScoreTag = "_2.NormalUserInitialize " //全てのIsuInitializeが終わってはじめて+1
	ScoreViewerInitialize      score.ScoreTag = "_3.ViewerInitialize     "
	ScoreViewerLoop            score.ScoreTag = "_4.ViewerLoop           "
	ScoreViewerDropout         score.ScoreTag = "_5.ViewerDropout        "
	ScoreRepairIsu             score.ScoreTag = "_6.RepairIsu            "
	ScorePostInfoCondition     score.ScoreTag = "_7.PostInfoCondition    "
	ScorePostWarningCondition  score.ScoreTag = "_8.PostWarningCondition "
	ScorePostCriticalCondition score.ScoreTag = "_9.PostCriticalCondition"
)

func SetScoreTags(scoreTable score.ScoreTable) {
	setScoreTag(scoreTable, ScoreStartBenchmark)
	setScoreTag(scoreTable, ScoreGraphGood)
	setScoreTag(scoreTable, ScoreGraphNormal)
	setScoreTag(scoreTable, ScoreGraphBad)
	setScoreTag(scoreTable, ScoreGraphWorst)
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
