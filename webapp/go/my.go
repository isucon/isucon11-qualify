package main

import (
	"sync"
	"time"
)

// isuUUID=>[]condition
var currentHourCond = map[string][]IsuCondition{}
var currentHourCondLock sync.Mutex

// isuUUID=>current hour
var currentHour = map[string]time.Time{}

var rowsToInsert []IsuCondition

// Returns rows to insert now
func addIsuConditionToPool(conds []IsuCondition) []IsuCondition {
	currentHourCondLock.Lock()
	defer currentHourCondLock.Unlock()

	for _, cond := range conds {
		hour := cond.Timestamp.Truncate(time.Hour)
		if hour == currentHour[cond.JIAIsuUUID] {
			if len(currentHourCond[cond.JIAIsuUUID]) > 10 {
				continue
			}
			currentHourCond[cond.JIAIsuUUID] = append(currentHourCond[cond.JIAIsuUUID], cond)
			// fmt.Printf("was same len %d\n", len(currentHourConditions[cond.JIAIsuUUID]))

		} else {
			rowsToInsert = append(rowsToInsert, currentHourCond[cond.JIAIsuUUID]...)
			currentHourCond[cond.JIAIsuUUID] = []IsuCondition{cond}
			currentHour[cond.JIAIsuUUID] = hour
		}
	}
	// fmt.Printf("was different len %d\n", len(rowsToInsert))
	if len(rowsToInsert) > 1000 {
		copy := rowsToInsert
		rowsToInsert = []IsuCondition{}
		return copy
	} else {
		return nil
	}
}
