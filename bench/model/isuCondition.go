package model

import "time"

//enum
type ConditionLevel int

const (
	ConditionLevelInfo ConditionLevel = iota
	ConditionLevelWarning
	ConditionLevelCritical
)

//TODO: メモリ節約の必要があるなら考える
type IsuCondition struct {
	Timestamp      time.Time      `json:"timestamp"`
	IsSitting      bool           `json:"is_sitting"`
	Condition      string         `json:"condition"`
	ConditionLevel ConditionLevel `json:"-"`
	Message        string         `json:"message"`
	Owner          *Isu
}
