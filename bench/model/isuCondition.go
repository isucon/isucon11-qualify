package model

//enum
type ConditionLevel int

const (
	ConditionLevelInfo ConditionLevel = iota
	ConditionLevelWarning
	ConditionLevelCritical
)

//TODO: メモリ節約の必要があるなら考える
type IsuCondition struct {
	TimestampUnix  int64          `json:"timestamp"`
	IsSitting      bool           `json:"is_sitting"`
	Condition      string         `json:"condition"`
	ConditionLevel ConditionLevel `json:"-"`
	Message        string         `json:"message"`
	Owner          *Isu
}
