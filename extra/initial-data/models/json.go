package models

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type JsonArray []*Json

func (j *JsonArray) Commit() error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("initialize.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}

type Json struct {
	JiaUserId   string      `json:"jia_user_id"`
	IsuListById IsuListById `json:"isu_list_by_id"`
}
type IsuListById map[string]JsonIsuInfo

type JsonIsuInfo struct {
	Name          string         `json:"name"`
	ImageFileHash string         `json:"image_file_hash"`
	Character     string         `json:"character"`
	Conditions    JsonConditions `json:"conditions"`
	CreatedAt     time.Time      `json:"created_at"`
}

func ToJsonIsuInfo(isu Isu, conditions JsonConditions) JsonIsuInfo {
	return JsonIsuInfo{
		isu.Name,
		fmt.Sprintf("%x", md5.Sum(isu.Image)),
		isu.Character,
		conditions,
		isu.CreatedAt,
	}
}

type JsonConditions struct {
	Info     []JsonCondition `json:"info"`
	Warning  []JsonCondition `json:"warning"`
	Critical []JsonCondition `json:"critical"`
}

type JsonCondition struct {
	Timestamp    int64     `json:"timestamp"`
	IsSitting    bool      `json:"is_sitting"`
	IsDirty      bool      `json:"is_dirty"`
	IsOverweight bool      `json:"is_overweight"`
	IsBroken     bool      `json:"is_broken"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

func (j *JsonConditions) AddCondition(condition Condition) error {
	jsonCondition := JsonCondition{
		condition.Timestamp.Unix(),
		condition.IsSitting,
		condition.IsDirty,
		condition.IsOverweight,
		condition.IsBroken,
		condition.Message,
		condition.CreatedAt,
	}
	switch condition.ConditionLevel() {
	case ConditionLevelInfo:
		j.Info = append(j.Info, jsonCondition)
	case ConditionLevelWarning:
		j.Warning = append(j.Warning, jsonCondition)
	case ConditionLevelCritical:
		j.Critical = append(j.Critical, jsonCondition)
	default:
		return fmt.Errorf("想定外のConditionLevelです。")
	}
	return nil
}
