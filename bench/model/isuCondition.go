package model

import (
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/isucon/isucon11-qualify/bench/model/eiya_redblacktree"
)

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

//TODO: メモリ節約の必要があるなら考える
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
	OwnerIsuUUID   string         `json:"owner_isu_uuid"`
}

//left < right
func (left *IsuCondition) Less(right *IsuCondition) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerIsuUUID < right.OwnerIsuUUID)
}

type IsuConditionCursor struct {
	TimestampUnix int64
	OwnerIsuUUID  string
}

//left < right
func (left *IsuConditionCursor) Less(right *IsuConditionCursor) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerIsuUUID < right.OwnerIsuUUID)
}

//left < right
func (left *IsuCondition) Less2(right *IsuConditionCursor) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerIsuUUID < right.OwnerIsuUUID)
}

//left < right
func (left *IsuConditionCursor) Less2(right *IsuCondition) bool {
	return left.TimestampUnix < right.TimestampUnix ||
		(left.TimestampUnix == right.TimestampUnix && left.OwnerIsuUUID < right.OwnerIsuUUID)
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

// UpperBound は IsuConditionArray から特定の時間「以下の」最新コンディションを指すイテレータを返す
func (ia *IsuConditionArray) UpperBound(filter ConditionLevel, targetTimestamp int64, targetOwnerIsuUUID string) IsuConditionArrayIterator {
	iter := ia.End(filter)
	if (iter.filter & ConditionLevelInfo) != 0 {
		iter.indexInfo = upperBoundIsuConditionIndex(iter.parent.Info, len(iter.parent.Info), targetTimestamp, targetOwnerIsuUUID)
	}
	if (iter.filter & ConditionLevelWarning) != 0 {
		iter.indexWarning = upperBoundIsuConditionIndex(iter.parent.Warning, len(iter.parent.Warning), targetTimestamp, targetOwnerIsuUUID)
	}
	if (iter.filter & ConditionLevelCritical) != 0 {
		iter.indexCritical = upperBoundIsuConditionIndex(iter.parent.Critical, len(iter.parent.Critical), targetTimestamp, targetOwnerIsuUUID)
	}
	return iter
}

// LowerBound は IsuConditionArray から特定の時間「より古い」最新コンディションを指すイテレータを返す
func (ia *IsuConditionArray) LowerBound(filter ConditionLevel, targetTimestamp int64, targetOwnerIsuUUID string) IsuConditionArrayIterator {
	iter := ia.End(filter)
	if (iter.filter & ConditionLevelInfo) != 0 {
		iter.indexInfo = lowerBoundIsuConditionIndex(iter.parent.Info, len(iter.parent.Info), targetTimestamp, targetOwnerIsuUUID)
	}
	if (iter.filter & ConditionLevelWarning) != 0 {
		iter.indexWarning = lowerBoundIsuConditionIndex(iter.parent.Warning, len(iter.parent.Warning), targetTimestamp, targetOwnerIsuUUID)
	}
	if (iter.filter & ConditionLevelCritical) != 0 {
		iter.indexCritical = lowerBoundIsuConditionIndex(iter.parent.Critical, len(iter.parent.Critical), targetTimestamp, targetOwnerIsuUUID)
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
func upperBoundIsuConditionIndex(base []IsuCondition, end int, targetTimestamp int64, targetOwnerIsuUUID string) int {
	//末尾の方にあることが分かっているので、末尾を固定要素ずつ線形探索 + 二分探索
	//assert end <= len(base)
	target := IsuConditionCursor{TimestampUnix: targetTimestamp, OwnerIsuUUID: targetOwnerIsuUUID}
	if end <= 0 {
		return end //要素が見つからない
	}
	//[0]が番兵になるかチェック
	if target.Less2(&base[0]) {
		return 0 //0がupperBound
	}

	//線形探索 ngがbase[ng] <= targetになるまで探索
	const defaultRange = 64
	ok := end
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

//baseはlessの昇順
func lowerBoundIsuConditionIndex(base []IsuCondition, end int, targetTimestamp int64, targetOwnerIsuUUID string) int {
	//末尾の方にあることが分かっているので、末尾を固定要素ずつ線形探索 + 二分探索
	//assert end <= len(base)
	target := IsuConditionCursor{TimestampUnix: targetTimestamp, OwnerIsuUUID: targetOwnerIsuUUID}
	if end <= 0 {
		return end //要素が見つからない
	}
	//[0]が番兵になるかチェック
	if !base[0].Less2(&target) {
		return 0 //0がupperBound
	}

	//線形探索 ngがbase[ng] <= targetになるまで探索
	const defaultRange = 64
	ok := end
	ng := end - defaultRange
	ng = (ng / defaultRange) * defaultRange //0未満になるのが嫌なので、defaultRangeの倍数にする
	for !base[ng].Less2(&target) {          //Timestampはunique仮定なので、<で良い（等価が見つかればそれで良し）
		ok = ng
		ng -= defaultRange
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

//TreeSet実装

//conditionをcreated atの大きい順で見る
type IsuConditionTreeSet struct {
	Info     *redblacktree.Tree
	Warning  *redblacktree.Tree
	Critical *redblacktree.Tree
}

//conditionをcreated atの大きい順で見る
type IsuConditionTreeSetIterator struct {
	filter   ConditionLevel
	info     eiya_redblacktree.Iterator
	warning  eiya_redblacktree.Iterator
	critical eiya_redblacktree.Iterator
	parent   *IsuConditionTreeSet
}

func NewIsuConditionTreeSet() IsuConditionTreeSet {
	comp := func(a, b interface{}) int {
		aAsserted := a.(*IsuCondition)
		bAsserted := b.(*IsuCondition)
		switch {
		case bAsserted.Less(aAsserted): //a > b
			return 1
		case aAsserted.Less(bAsserted):
			return -1
		default:
			return 0
		}
	}
	return IsuConditionTreeSet{
		Info:     redblacktree.NewWith(comp),
		Warning:  redblacktree.NewWith(comp),
		Critical: redblacktree.NewWith(comp),
	}
}

func (is *IsuConditionTreeSet) Add(cond *IsuCondition) {
	switch cond.ConditionLevel {
	case ConditionLevelInfo:
		is.Info.Put(cond, struct{}{})
	case ConditionLevelWarning:
		is.Warning.Put(cond, struct{}{})
	case ConditionLevelCritical:
		is.Critical.Put(cond, struct{}{})
	}
}

func (ia *IsuConditionTreeSet) End(filter ConditionLevel) IsuConditionTreeSetIterator {
	iter := IsuConditionTreeSetIterator{
		filter:   filter,
		info:     eiya_redblacktree.NewIterator(ia.Info),
		warning:  eiya_redblacktree.NewIterator(ia.Warning),
		critical: eiya_redblacktree.NewIterator(ia.Critical),
		parent:   ia,
	}
	iter.info.End()
	iter.warning.End()
	iter.critical.End()
	return iter
}

func (ia *IsuConditionTreeSet) Back() *IsuCondition {
	iter := ia.End(ConditionLevelInfo | ConditionLevelWarning | ConditionLevelCritical)
	return iter.Prev()
}

func (ia *IsuConditionTreeSet) LowerBound(filter ConditionLevel, targetTimestamp int64, targetOwnerIsuUUID string) IsuConditionTreeSetIterator {
	iter := ia.End(filter)
	cursor := &IsuCondition{TimestampUnix: targetTimestamp, OwnerIsuUUID: targetOwnerIsuUUID}
	if (filter & ConditionLevelInfo) != 0 {
		node, found := ia.Info.Ceiling(cursor)
		if found {
			iter.info = eiya_redblacktree.NewIteratorWithNode(ia.Info, node)
		} else {
			iter.info.End()
		}
	}
	if (filter & ConditionLevelWarning) != 0 {
		node, found := ia.Warning.Ceiling(cursor)
		if found {
			iter.warning = eiya_redblacktree.NewIteratorWithNode(ia.Warning, node)
		} else {
			iter.warning.End()
		}
	}
	if (filter & ConditionLevelCritical) != 0 {
		node, found := ia.Critical.Ceiling(cursor)
		if found {
			iter.critical = eiya_redblacktree.NewIteratorWithNode(ia.Critical, node)
		} else {
			iter.critical.End()
		}
	}
	return iter
}

//return: nil:もう要素がない
func (iter *IsuConditionTreeSetIterator) Prev() *IsuCondition {
	maxType := ConditionLevelNone
	var max *IsuCondition
	if (iter.filter & ConditionLevelInfo) != 0 {
		prevOK := iter.info.Prev()
		if prevOK && (max == nil || max.Less(iter.info.Key().(*IsuCondition))) {
			maxType = ConditionLevelInfo
			max = iter.info.Key().(*IsuCondition)
		}
		iter.info.Next()
	}
	if (iter.filter & ConditionLevelWarning) != 0 {
		prevOK := iter.warning.Prev()
		if prevOK && (max == nil || max.Less(iter.warning.Key().(*IsuCondition))) {
			maxType = ConditionLevelWarning
			max = iter.warning.Key().(*IsuCondition)
		}
		iter.warning.Next()
	}
	if (iter.filter & ConditionLevelCritical) != 0 {
		prevOK := iter.critical.Prev()
		if prevOK && (max == nil || max.Less(iter.critical.Key().(*IsuCondition))) {
			maxType = ConditionLevelCritical
			max = iter.critical.Key().(*IsuCondition)
		}
		iter.critical.Next()
	}

	switch maxType {
	case ConditionLevelInfo:
		iter.info.Prev()
	case ConditionLevelWarning:
		iter.warning.Prev()
	case ConditionLevelCritical:
		iter.critical.Prev()
	}
	return max
}
