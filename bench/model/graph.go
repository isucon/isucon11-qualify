package model

const (
	// graph の1要素 (1h) を構成するコンディション数のボーダー値
	// この値を下回った分だけ減点処理される
	missingDataBorder = 50

	scoreConditionLevelInfo     = 3
	scoreConditionLevelWarning  = 2
	scoreConditionLevelCritical = 1
)

var (
	scorePerCondition = map[string]int{
		"is_dirty":      -1,
		"is_overweight": -1,
		"is_broken":     -5,
	}
)

// Graph Model は verifyGraph にて利用される。
// getGraph のレスポンスボディのうち condition_timestamps を元にモデルを組み立て、
// モデルを用いて getGraph のレスポンスボディの他フィールドが適切な値かどうかを検証する。
type Graph struct {
	conditions []*IsuCondition

	score      int
	percentage GraphDetail
}

type GraphDetail struct {
	sitting      int
	isBroken     int
	isDirty      int
	isOverweight int
}

func NewGraph(c []*IsuCondition) Graph {
	graph := Graph{conditions: c}
	graph.calculate()
	return graph
}

// g.condition を元に他フィールドを埋めるメソッド
func (g *Graph) calculate() {
	rawScore := 0
	sittingCount := 0
	brokenCount := 0
	overweightCount := 0
	dirtyCount := 0
	for _, c := range g.conditions {
		warnCount := 0
		if c.IsSitting {
			sittingCount++
		}
		if c.IsDirty {
			warnCount++
			dirtyCount++
		}
		if c.IsOverweight {
			warnCount++
			overweightCount++
		}
		if c.IsBroken {
			warnCount++
			brokenCount++
		}
		switch warnCount {
		case 0:
			rawScore += scoreConditionLevelInfo
		case 1, 2:
			rawScore += scoreConditionLevelWarning
		case 3:
			rawScore += scoreConditionLevelCritical
		}
	}
	rawScore = rawScore * 100 / 3
	g.score = rawScore / len(g.conditions)
	g.percentage.sitting = sittingCount * 100 / len(g.conditions)
	g.percentage.isBroken = brokenCount * 100 / len(g.conditions)
	g.percentage.isOverweight = overweightCount * 100 / len(g.conditions)
	g.percentage.isDirty = dirtyCount * 100 / len(g.conditions)
}

func (g Graph) Match(
	score int,
	sittingPercentage int,
	isBrokenPercentage int,
	isDirtyPercentage int,
	isOverweightPercentage int,
) bool {
	return score == g.score &&
		sittingPercentage == g.percentage.sitting &&
		isBrokenPercentage == g.percentage.isBroken &&
		isDirtyPercentage == g.percentage.isDirty &&
		isOverweightPercentage == g.percentage.isOverweight
}
