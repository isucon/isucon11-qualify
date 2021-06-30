package models

import (
	"fmt"
	"log"
	"time"

	"github.com/isucon/isucon11-qualify/extra/initial-data/graph"
	"github.com/isucon/isucon11-qualify/extra/initial-data/random"
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

func NewCondition(isu Isu) Condition {
	t := random.TimeAfterArg(isu.CreatedAt)
	isSitting, isDirty, isOverweigh, isBroken := random.Condition()
	return Condition{
		isu,
		t,
		isSitting,
		isDirty,
		isOverweigh,
		isBroken,
		random.MessageWithCondition(isSitting, isDirty, isOverweigh, isBroken),
		t,
	}
}

// MEMO: random.baseTime を超えた時間が入る可能性がある
func NewConditionFromLastCondition(c Condition, durationMinute int) Condition {
	c.Timestamp = c.Timestamp.Add(time.Duration(durationMinute) * time.Minute)
	c.CreatedAt = c.Timestamp
	c.IsSitting = random.IsSittingFromLastCondition(c.IsSitting)
	c.IsDirty = random.IsDirtyFromLastCondition(c.IsDirty)
	c.IsOverweight = random.IsOverweightFromLastCondition(c.IsOverweight)
	c.IsBroken = random.IsBrokenFromLastCondition(c.IsBroken)
	return c
}

func (c Condition) Create() error {
	// INSERT INTO isu_condition
	condition := fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", c.IsDirty, c.IsOverweight, c.IsBroken)
	if _, err := db.Exec("INSERT INTO isu_condition VALUES (?,?,?,?,?,?)",
		c.Isu.JIAIsuUUID, c.Timestamp, c.IsSitting, condition, c.Message, c.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	// INSERT INTO graph
	if err := graph.UpdateGraph(db, c.Isu.JIAIsuUUID, c.CreatedAt); err != nil {
		log.Fatal(err)
	}
	return nil
}
