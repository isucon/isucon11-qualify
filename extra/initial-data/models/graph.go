package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Graph struct {
	Isu Isu
}

func NewGraph(isu Isu) *Graph {
	return &Graph{
		isu,
	}
}

func (g Graph) CreateWithCondition(condition Condition) error {
	url := fmt.Sprintf("%sapi/isu/%s/condition", apiUrl, g.Isu.JIAIsuUUID)
	c := struct {
		IsSitting bool   `json:"is_sitting"`
		Condition string `json:"condition"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"` //Format("2006-01-02 15:04:05 -0700")
	}{
		condition.IsSitting,
		fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", condition.IsDirty, condition.IsOverweight, condition.IsBroken),
		condition.Message,
		condition.Timestamp.Format(time.RFC3339),
	}
	data, err := json.Marshal(&c)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	if _, err := client.Do(req); err != nil {
		return err
	}
	return nil
}
