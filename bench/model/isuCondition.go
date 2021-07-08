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
	return left.TimestampUnix < right.TimestampUnix &&
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

func (ia *IsuConditionArray) End(filter ConditionLevel) IsuConditionIterator {
	return IsuConditionIterator{
		filter:        filter,
		indexInfo:     len(ia.Info),
		indexWarning:  len(ia.Warning),
		indexCritical: len(ia.Critical),
		parent:        ia,
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
