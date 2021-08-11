package models

import (
	"fmt"
	"time"

	"github.com/isucon/isucon11-qualify/bench/random"
)

type Condition struct {
	Isu          Isu
	Timestamp    time.Time
	IsSitting    bool
	IsDirty      bool
	IsOverweight bool
	IsBroken     bool
	Message      string
	CreatedAt    time.Time
}
type ConditionLevel int

const (
	ConditionLevelInfo     ConditionLevel = 1
	ConditionLevelWarning  ConditionLevel = 2
	ConditionLevelCritical ConditionLevel = 4
)

func NewCondition(isu Isu) Condition {
	t := isu.CreatedAt.Add(time.Minute) // 初回 condition は ISU が作成された時間 + 1分後
	isSitting, isDirty, isOverweigh, isBroken := random.Condition()
	return Condition{
		isu,
		t,
		isSitting,
		isDirty,
		isOverweigh,
		isBroken,
		random.MessageWithCondition(isDirty, isOverweigh, isBroken, isu.CharacterId),
		t,
	}
}

// MEMO: random.baseTime を超えた時間が入る可能性がある
func NewConditionFromLastCondition(c Condition, durationMinute int) Condition {
	c.Timestamp = c.Timestamp.Add(time.Duration(durationMinute) * time.Minute) // 前回 condition を送信した時間の durationMinute 後
	c.CreatedAt = c.Timestamp
	c.IsSitting = random.IsSittingFromLastCondition(c.IsSitting)
	c.IsDirty = random.IsDirtyFromLastCondition(c.IsDirty)
	c.IsOverweight = random.IsOverweightFromLastCondition(c.IsOverweight)
	c.IsBroken = random.IsBrokenFromLastCondition(c.IsBroken)
	c.Message = random.MessageWithCondition(c.IsDirty, c.IsOverweight, c.IsBroken, c.Isu.CharacterId)
	return c
}

func (c Condition) Create() error {
	// INSERT INTO isu_condition
	condition := fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", c.IsDirty, c.IsOverweight, c.IsBroken)
	if _, err := db.Exec("INSERT INTO isu_condition(`jia_isu_uuid`, `timestamp`, `is_sitting`, `condition`, `message`, `created_at`) VALUES (?,?,?,?,?,?)",
		c.Isu.JIAIsuUUID, c.Timestamp, c.IsSitting, condition, c.Message, c.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert isu_condition: %w", err)
	}

	return nil
}

func (c Condition) ConditionLevel() ConditionLevel {
	warnCount := 0
	if c.IsDirty {
		warnCount += 1
	}
	if c.IsOverweight {
		warnCount += 1
	}
	if c.IsBroken {
		warnCount += 1
	}
	if warnCount == 0 {
		return ConditionLevelInfo
	} else if warnCount == 1 || warnCount == 2 {
		return ConditionLevelWarning
	} else {
		return ConditionLevelCritical
	}
}
