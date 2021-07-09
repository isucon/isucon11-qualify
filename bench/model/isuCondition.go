package model

//enum
type ConditionLevel int

const (
	ConditionLevelNone     ConditionLevel = 0
	ConditionLevelInfo     ConditionLevel = 1
	ConditionLevelWarning  ConditionLevel = 2
	ConditionLevelCritical ConditionLevel = 4
)

//TODO: メモリ節約の必要があるなら考える
type IsuCondition struct {
	StateChange   IsuStateChange
	TimestampUnix int64 `json:"timestamp"`
	IsSitting     bool  `json:"is_sitting"`
	//Condition      string         `json:"condition"`
	IsDirty        bool
	IsOverweight   bool
	IsBroken       bool
	ConditionLevel ConditionLevel `json:"-"`
	Message        string         `json:"message"`
	OwnerID        string
	//	Owner          *Isu
}

//left < right
func (left *IsuCondition) Less(right *IsuCondition) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerID < right.OwnerID)
}

type IsuConditionCursor struct {
	TimestampUnix int64
	OwnerID       string
}

//left < right
func (left *IsuConditionCursor) Less(right *IsuConditionCursor) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerID < right.OwnerID)
}

//left < right
func (left *IsuCondition) Less2(right *IsuConditionCursor) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerID < right.OwnerID)
}

//left < right
func (left *IsuConditionCursor) Less2(right *IsuCondition) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerID < right.OwnerID)
}

//conditionをcreated at順で見る
type IsuConditionArray struct {
	Info     []IsuCondition
	Warning  []IsuCondition
	Critical []IsuCondition
}

//conditionをcreated atの大きい順で見る
type IsuConditionIterator struct {
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

func (ia *IsuConditionArray) End(filter ConditionLevel) IsuConditionIterator {
	return IsuConditionIterator{
		filter:        filter,
		indexInfo:     len(ia.Info),
		indexWarning:  len(ia.Warning),
		indexCritical: len(ia.Critical),
		parent:        ia,
	}
}

func (iter *IsuConditionIterator) UpperBoundIsuConditionIndex(targetTimestamp int64, targetIsuUUID string) {
	if (iter.filter & ConditionLevelInfo) != 0 {
		iter.indexInfo = upperBoundIsuConditionIndex(iter.parent.Info, len(iter.parent.Info), targetTimestamp, targetIsuUUID)
	}
	if (iter.filter & ConditionLevelWarning) != 0 {
		iter.indexWarning = upperBoundIsuConditionIndex(iter.parent.Warning, len(iter.parent.Warning), targetTimestamp, targetIsuUUID)
	}
	if (iter.filter & ConditionLevelCritical) != 0 {
		iter.indexCritical = upperBoundIsuConditionIndex(iter.parent.Critical, len(iter.parent.Critical), targetTimestamp, targetIsuUUID)
	}
}

//return: nil:もう要素がない
func (iter *IsuConditionIterator) Prev() *IsuCondition {
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
func upperBoundIsuConditionIndex(base []IsuCondition, end int, targetTimestamp int64, targetIsuUUID string) int {
	//末尾の方にあることが分かっているので、末尾を固定要素ずつ線形探索 + 二分探索
	//assert end <= len(base)
	target := IsuConditionCursor{TimestampUnix: targetTimestamp, OwnerID: targetIsuUUID}
	if end <= 0 {
		return end //要素が見つからない
	}
	//[0]が番兵になるかチェック
	if target.Less2(&base[0]) {
		return 0 //0がupperBound
	}

	//線形探索 ngがbase[ng] <= targetになるまで探索
	const defaultRange = 64
	ok := end - 1
	ng := end - defaultRange
	ng = (ng / defaultRange) * defaultRange //0未満になるのが嫌なので、defaultRangeの倍数にする
	for target.Less2(&base[ng]) {           //Timestampはunique仮定なので、<で良い（等価が見つかればそれで良し）
		ok = ng
		ng -= defaultRange
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
