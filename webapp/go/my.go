package main

import (
	"sync"
	"time"
)

// isuUUID=>[]condition
var currentHourCond = map[string][]IsuCondition{}
var currentHourCondLock sync.Mutex
var latestCond = map[string]IsuCondition{}
var latestCondLock sync.Mutex

// isuUUID=>current hour
var currentHour = map[string]time.Time{}

var rowsToInsert []IsuCondition

func refreshLatestCondition(conds []IsuCondition) {
	latestCondLock.Lock()
	defer latestCondLock.Unlock()
	for _, cond := range conds {
		if cond.Timestamp.After(latestCond[cond.JIAIsuUUID].Timestamp) {
			latestCond[cond.JIAIsuUUID] = cond
		}
	}
}

var recentLatestCondition time.Time
var cachedLatestConditionResponse = []IsuCondition{}

func getLatestConditions() []IsuCondition {
	if time.Now().Before(recentLatestCondition.Add(500 * time.Millisecond)) {
		return cachedLatestConditionResponse
	}

	latestCondLock.Lock()
	defer latestCondLock.Unlock()

	newResponse := []IsuCondition{}
	for _, cond := range latestCond {
		newResponse = append(newResponse, cond)
	}
	cachedLatestConditionResponse = newResponse
	return cachedLatestConditionResponse
}

func getLatestConditionsAsMap(userId string) map[string]IsuCondition {
	latestCondLock.Lock()
	defer latestCondLock.Unlock()

	result := map[string]IsuCondition{}
	for isuUUID, cond := range latestCond {
		if cond.UserId == userId {
			result[isuUUID] = cond
		}
	}
	return result
}

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

		} else {
			rowsToInsert = append(rowsToInsert, currentHourCond[cond.JIAIsuUUID]...)
			currentHourCond[cond.JIAIsuUUID] = []IsuCondition{cond}
			currentHour[cond.JIAIsuUUID] = hour
		}
	}
	if len(rowsToInsert) > 1000 {
		copy := rowsToInsert
		rowsToInsert = []IsuCondition{}
		return copy
	} else {
		return nil
	}
}

var isusCacheLock sync.Mutex

var isusCache = map[string]Isu{}

func getIsu(uuid string) (Isu, bool) {
	isusCacheLock.Lock()
	defer isusCacheLock.Unlock()
	isu, ok := isusCache[uuid]
	return isu, ok
}

func addIsu(newIsu Isu) {
	isusCacheLock.Lock()
	defer isusCacheLock.Unlock()
	isusCache[newIsu.JIAIsuUUID] = newIsu
}

func loadLatestConditionFromDb() error {
	latestCond = map[string]IsuCondition{}

	conds := []IsuCondition{}

	err := db.Select(&conds,
		"SELECT a.character, a.id as isu_id, a.jia_user_id, b.id, b.message, b.timestamp, b.condition, b.jia_isu_uuid, b.is_sitting, b.created_at FROM isu a INNER JOIN isu_condition b ON a.jia_isu_uuid = b.jia_isu_uuid WHERE b.timestamp = (SELECT timestamp FROM isu_condition WHERE jia_isu_uuid = a.jia_isu_uuid ORDER BY timestamp DESC limit 1)")
	if err != nil {
		return err
	}

	for _, cond := range conds {
		latestCond[cond.JIAIsuUUID] = cond
	}
	return nil
}

var users = map[string]bool{}
var usersLock = sync.Mutex{}

func checkUser(userId string) (bool, error) {
	usersLock.Lock()
	defer usersLock.Unlock()

	if val, ok := users[userId]; ok {
		return val, nil
	} else {
		count := 0
		err := db.Get(&count, "SELECT COUNT(*) FROM `user` WHERE `jia_user_id` = ?",
			userId)
		if err != nil {
			return false, err
		}

		users[userId] = count > 0
		return users[userId], nil
	}
}
