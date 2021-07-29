package model

const (
	// graph の1要素 (1h) を構成するコンディション数のボーダー値
	// この値を下回った分だけ減点処理される
	missingDataBorder = 50
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

	score   int
	sitting int
	detail  GraphDetail
}

type GraphDetail struct {
	isBroken     int
	isDirty      int
	isOverweight int
	missingData  int
}

func NewGraph(c []*IsuCondition) *Graph {
	graph := &Graph{conditions: c}
	graph.calculate()
	return graph
}

// g.condition を元に他フィールドを埋めるメソッド
func (g *Graph) calculate() {
	sittingCount := 0
	for _, log := range g.conditions {
		if log.IsSitting {
			sittingCount++
		}
	}
	g.sitting = sittingCount * 100 / len(g.conditions)

	//score&detail
	g.score = 100
	//condition要因の減点
	for _, c := range g.conditions {
		//trueになっているものは減点
		if c.IsDirty {
			g.score += scorePerCondition["is_dirty"]
			g.detail.isDirty += scorePerCondition["is_dirty"]
		}
		if c.IsOverweight {
			g.score += scorePerCondition["is_overweight"]
			g.detail.isOverweight += scorePerCondition["is_overweight"]
		}
		if c.IsBroken {
			g.score += scorePerCondition["is_broken"]
			g.detail.isBroken += scorePerCondition["is_broken"]
		}
	}
	//個数減点
	if len(g.conditions) < missingDataBorder {
		minus := -(missingDataBorder - len(g.conditions)) * 2
		g.score += minus
		g.detail.missingData = minus
	}
	if g.score < 0 {
		g.score = 0
	}
}

// getters

func (g Graph) Score() int {
	return g.score
}
func (g Graph) Sitting() int {
	return g.sitting
}
func (g Graph) IsBroken() int {
	return g.detail.isBroken
}
func (g Graph) IsDirty() int {
	return g.detail.isDirty
}
func (g Graph) IsOverweight() int {
	return g.detail.isOverweight
}
func (g Graph) MissingData() int {
	return g.detail.missingData
}
