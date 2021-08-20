package model

//enum
type ConditionLevel int

const (
	ConditionLevelNone     ConditionLevel = 0
	ConditionLevelInfo     ConditionLevel = 1
	ConditionLevelWarning  ConditionLevel = 2
	ConditionLevelCritical ConditionLevel = 4
)

func (cl ConditionLevel) Equal(conditionLevel string) bool {
	if (cl == ConditionLevelInfo && conditionLevel == "info") ||
		(cl == ConditionLevelWarning && conditionLevel == "warning") ||
		(cl == ConditionLevelCritical && conditionLevel == "critical") {
		return true
	}
	return false
}

type IsuCondition struct {
	StateChange   IsuStateChange
	TimestampUnix int64 `json:"timestamp"`
	IsSitting     bool  `json:"is_sitting"`
	//Condition      string         `json:"condition"`
	IsDirty        bool           `json:"is_dirty"`
	IsOverweight   bool           `json:"is_overweight"`
	IsBroken       bool           `json:"is_broken"`
	ConditionLevel ConditionLevel `json:"condition_level"`
	Message        string         `json:"message"`

	ReadTime int64 `json:"-"` // GET /api/condition/:id や GET /api/isu/:id/graph で読まれた実際の時間(仮想時間ではない)
}

//left < right
func (left *IsuCondition) Less(right *IsuCondition) bool {
	return left.TimestampUnix < right.TimestampUnix
}

type IsuConditionCursor struct {
	TimestampUnix int64
}

//left < right
func (left *IsuConditionCursor) Less(right *IsuConditionCursor) bool {
	return left.TimestampUnix < right.TimestampUnix
}

//left < right
func (left *IsuCondition) Less2(right *IsuConditionCursor) bool {
	return left.TimestampUnix < right.TimestampUnix
}

//left < right
func (left *IsuConditionCursor) Less2(right *IsuCondition) bool {
	return left.TimestampUnix < right.TimestampUnix
}

//conditionをcreated atの大きい順で見る
type IsuConditionIterator interface {
	Prev() *IsuCondition
}

//Array実装

//conditionをcreated atの大きい順で見る
type IsuConditionArray struct {
	Info     []IsuCondition
	Warning  []IsuCondition
	Critical []IsuCondition
}

//conditionをcreated atの大きい順で見る
type IsuConditionArrayIterator struct {
	filter        ConditionLevel
	indexInfo     int
	indexWarning  int
	indexCritical int
	parent        *IsuConditionArray
}

func NewIsuConditionArray() IsuConditionArray {
	return IsuConditionArray{
		Info:     []IsuCondition{},
		Warning:  []IsuCondition{},
		Critical: []IsuCondition{},
	}
}

func (ia *IsuConditionArray) Add(cond *IsuCondition) {
	switch cond.ConditionLevel {
	case ConditionLevelInfo:
		ia.Info = append(ia.Info, *cond)
	case ConditionLevelWarning:
		ia.Warning = append(ia.Warning, *cond)
	case ConditionLevelCritical:
		ia.Critical = append(ia.Critical, *cond)
	}
}

func (ia *IsuConditionArray) End(filter ConditionLevel) IsuConditionArrayIterator {
	return IsuConditionArrayIterator{
		filter:        filter,
		indexInfo:     len(ia.Info),
		indexWarning:  len(ia.Warning),
		indexCritical: len(ia.Critical),
		parent:        ia,
	}
}

func (ia *IsuConditionArray) Back() *IsuCondition {
	iter := ia.End(ConditionLevelInfo | ConditionLevelWarning | ConditionLevelCritical)
	return iter.Prev()
}

// IsuConditionArrayは、後ろの方が新しい
// UpperBound は IsuConditionArray から特定の時間「より新しい」最も古い(手前の)コンディションを指すイテレータを返す
func (ia *IsuConditionArray) UpperBound(filter ConditionLevel, targetTimestamp int64) IsuConditionArrayIterator {
	iter := ia.End(filter)
	if (iter.filter & ConditionLevelInfo) != 0 {
		iter.indexInfo = upperBoundIsuConditionIndex(iter.parent.Info, len(iter.parent.Info), targetTimestamp)
	}
	if (iter.filter & ConditionLevelWarning) != 0 {
		iter.indexWarning = upperBoundIsuConditionIndex(iter.parent.Warning, len(iter.parent.Warning), targetTimestamp)
	}
	if (iter.filter & ConditionLevelCritical) != 0 {
		iter.indexCritical = upperBoundIsuConditionIndex(iter.parent.Critical, len(iter.parent.Critical), targetTimestamp)
	}
	return iter
}

// IsuConditionArrayは、後ろの方が新しい
// LowerBound は IsuConditionArray から特定の時間「以上の」最も古い(手前の)コンディションを指すイテレータを返す
func (ia *IsuConditionArray) LowerBound(filter ConditionLevel, targetTimestamp int64) IsuConditionArrayIterator {
	iter := ia.End(filter)
	if (iter.filter & ConditionLevelInfo) != 0 {
		iter.indexInfo = lowerBoundIsuConditionIndex(iter.parent.Info, len(iter.parent.Info), targetTimestamp)
	}
	if (iter.filter & ConditionLevelWarning) != 0 {
		iter.indexWarning = lowerBoundIsuConditionIndex(iter.parent.Warning, len(iter.parent.Warning), targetTimestamp)
	}
	if (iter.filter & ConditionLevelCritical) != 0 {
		iter.indexCritical = lowerBoundIsuConditionIndex(iter.parent.Critical, len(iter.parent.Critical), targetTimestamp)
	}
	return iter
}

//return: nil:もう要素がない
func (iter *IsuConditionArrayIterator) Prev() *IsuCondition {
	maxType := ConditionLevelNone
	var max *IsuCondition
	if (iter.filter&ConditionLevelInfo) != 0 && iter.indexInfo != 0 {
		if max == nil || max.Less(&iter.parent.Info[iter.indexInfo-1]) {
			maxType = ConditionLevelInfo
			max = &iter.parent.Info[iter.indexInfo-1]
		}
	}
	if (iter.filter&ConditionLevelWarning) != 0 && iter.indexWarning != 0 {
		if max == nil || max.Less(&iter.parent.Warning[iter.indexWarning-1]) {
			maxType = ConditionLevelWarning
			max = &iter.parent.Warning[iter.indexWarning-1]
		}
	}
	if (iter.filter&ConditionLevelCritical) != 0 && iter.indexCritical != 0 {
		if max == nil || max.Less(&iter.parent.Critical[iter.indexCritical-1]) {
			maxType = ConditionLevelCritical
			max = &iter.parent.Critical[iter.indexCritical-1]
		}
	}

	switch maxType {
	case ConditionLevelInfo:
		iter.indexInfo--
	case ConditionLevelWarning:
		iter.indexWarning--
	case ConditionLevelCritical:
		iter.indexCritical--
	}
	return max
}

//baseはlessの昇順
//「より大きい」を返す（C++と同じ）
func upperBoundIsuConditionIndex(base []IsuCondition, end int, targetTimestamp int64) int {
	//末尾の方にあることが分かっているので、末尾を固定要素ずつ線形探索 + 二分探索
	//assert end <= len(base)
	target := IsuConditionCursor{TimestampUnix: targetTimestamp}
	if end <= 0 {
		return end //要素が見つからない
	}
	//[0]が番兵になるかチェック
	if target.Less2(&base[0]) {
		return 0 //0がupperBound
	}

	//線形探索 ngがbase[ng] <= targetになるまで探索
	searchRange := 64
	ok := end
	ng := end - searchRange
	if ng < 0 {
		ng = 0
	}
	for target.Less2(&base[ng]) { //Timestampはunique仮定なので、<で良い（等価が見つかればそれで良し）
		ok = ng
		ng -= searchRange
		searchRange *= 2
		if ng < 0 {
			ng = 0
		}
	}

	//答えは(ng, ok]内にあるはずなので、二分探索
	for ok-ng > 1 {
		mid := (ok + ng) / 2
		if target.Less2(&base[mid]) {
			ok = mid
		} else {
			ng = mid
		}
	}

	return ok
}

//baseはlessの昇順
//「以上」を返す（C++と同じ）
func lowerBoundIsuConditionIndex(base []IsuCondition, end int, targetTimestamp int64) int {
	//末尾の方にあることが分かっているので、末尾を固定要素ずつ線形探索 + 二分探索
	//assert end <= len(base)
	target := IsuConditionCursor{TimestampUnix: targetTimestamp}
	if end <= 0 {
		return end //要素が見つからない
	}
	//[0]が番兵になるかチェック
	if !base[0].Less2(&target) {
		return 0 //0がupperBound
	}

	//線形探索 ngがbase[ng] <= targetになるまで探索
	searchRange := 64
	ok := end
	ng := end - searchRange
	if ng < 0 {
		ng = 0
	}
	for !base[ng].Less2(&target) { //Timestampはunique仮定なので、<で良い（等価が見つかればそれで良し）
		ok = ng
		ng -= searchRange
		searchRange *= 2
		if ng < 0 {
			ng = 0
		}
	}

	//答えは(ng, ok]内にあるはずなので、二分探索
	for ok-ng > 1 {
		mid := (ok + ng) / 2
		if !base[mid].Less2(&target) {
			ok = mid
		} else {
			ng = mid
		}
	}

	return ok
}

func (cond *IsuCondition) ConditionString() string {
	if cond.IsDirty {
		if cond.IsOverweight {
			if cond.IsBroken {
				return "is_dirty=true,is_overweight=true,is_broken=true"
			} else {
				return "is_dirty=true,is_overweight=true,is_broken=false"
			}
		} else {
			if cond.IsBroken {
				return "is_dirty=true,is_overweight=false,is_broken=true"
			} else {
				return "is_dirty=true,is_overweight=false,is_broken=false"
			}
		}
	} else {
		if cond.IsOverweight {
			if cond.IsBroken {
				return "is_dirty=false,is_overweight=true,is_broken=true"
			} else {
				return "is_dirty=false,is_overweight=true,is_broken=false"
			}
		} else {
			if cond.IsBroken {
				return "is_dirty=false,is_overweight=false,is_broken=true"
			} else {
				return "is_dirty=false,is_overweight=false,is_broken=false"
			}
		}
	}
}
