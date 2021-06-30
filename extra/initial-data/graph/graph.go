package graph

// copy from github.com/isucon/isucon11-qualify/webapp/go/main.go

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type IsuLog struct {
	JIAIsuUUID string    `db:"jia_isu_uuid" json:"jia_isu_uuid"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"`
	IsSitting  bool      `db:"is_sitting" json:"is_sitting"`
	Condition  string    `db:"condition" json:"condition"`
	Message    string    `db:"message" json:"message"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type GraphData struct {
	Score   int            `json:"score"`
	Sitting int            `json:"sitting"`
	Detail  map[string]int `json:"detail"`
}

var scorePerCondition = map[string]int{
	"is_dirty":      -1,
	"is_overweight": -1,
	"is_broken":     -5,
}

func UpdateGraph(x sqlx.Ext, jiaIsuUUID string, updatedAt time.Time) error {
	// IsuLogを一時間ごとの区切りに分け、区切りごとにスコアを計算する
	isuLogCluster := []IsuLog{} // 一時間ごとの纏まり
	var tmpIsuLog IsuLog
	valuesForUpdate := []interface{}{} //5個1組、更新するgraphの各行のデータ
	rows, err := x.Queryx("SELECT * FROM `isu_log` WHERE `jia_isu_uuid` = ? ORDER BY `timestamp` ASC", jiaIsuUUID)
	if err != nil {
		return err
	}
	//一時間ごとに区切る
	var startTime time.Time
	for rows.Next() {
		err = rows.StructScan(&tmpIsuLog)
		if err != nil {
			return err
		}
		tmpTime := truncateAfterHours(tmpIsuLog.Timestamp)
		if startTime != tmpTime {
			if len(isuLogCluster) > 0 {
				//tmpTimeは次の一時間なので、それ以外を使ってスコア計算
				data, err := calculateGraphData(isuLogCluster)
				if err != nil {
					return fmt.Errorf("failed to calculate graph: %v", err)
				}
				valuesForUpdate = append(valuesForUpdate, jiaIsuUUID, startTime, data, updatedAt, updatedAt)
			}

			//次の一時間の探索
			startTime = tmpTime
			isuLogCluster = []IsuLog{}
		}
		isuLogCluster = append(isuLogCluster, tmpIsuLog)
	}
	if len(isuLogCluster) > 0 {
		//最後の一時間分
		data, err := calculateGraphData(isuLogCluster)
		if err != nil {
			return fmt.Errorf("failed to calculate graph: %v", err)
		}
		valuesForUpdate = append(valuesForUpdate, jiaIsuUUID, startTime, data, updatedAt, updatedAt)
	}

	//insert or update
	params := strings.Repeat("(?,?,?,?,?),", len(valuesForUpdate)/5)
	params = params[:len(params)-1]
	_, err = x.Exec("INSERT INTO `graph` (`jia_isu_uuid`, `start_at`, `data`, `created_at`, `updated_at`) VALUES "+
		params+
		"	ON DUPLICATE KEY UPDATE `data` = VALUES(`data`), `updated_at` = VALUES(`updated_at`)",
		valuesForUpdate...,
	)
	if err != nil {
		return err
	}

	return nil
}

//分以下を切り捨て、一時間単位にする関数
func truncateAfterHours(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

//スコア計算をする関数
func calculateGraphData(isuLogCluster []IsuLog) ([]byte, error) {
	graph := &GraphData{}

	//sitting
	sittingCount := 0
	for _, log := range isuLogCluster {
		if log.IsSitting {
			sittingCount++
		}
	}
	graph.Sitting = sittingCount * 100 / len(isuLogCluster)

	//score&detail
	graph.Score = 100
	//condition要因の減点
	graph.Detail = map[string]int{}
	for key := range scorePerCondition {
		graph.Detail[key] = 0
	}
	for _, log := range isuLogCluster {
		conditions := map[string]bool{}
		//DB上にある is_dirty=true/false,is_overweight=true/false,... 形式のデータを
		//map[string]bool形式に変換
		for _, cond := range strings.Split(log.Condition, ",") {
			keyValue := strings.Split(cond, "=")
			if len(keyValue) != 2 {
				continue //形式に従っていないものは無視
			}
			conditions[keyValue[0]] = (keyValue[1] != "false")
		}

		//trueになっているものは減点
		for key, enabled := range conditions {
			if enabled {
				score, ok := scorePerCondition[key]
				if ok {
					graph.Score += score
					graph.Detail[key] += score
				}
			}
		}
	}
	//スコアに影響がないDetailを削除
	for key := range scorePerCondition {
		if graph.Detail[key] == 0 {
			delete(graph.Detail, key)
		}
	}
	//個数減点
	if len(isuLogCluster) < 50 {
		minus := -(50 - len(isuLogCluster)) * 2
		graph.Score += minus
		graph.Detail["missing_data"] = minus
	}
	if graph.Score < 0 {
		graph.Score = 0
	}

	//JSONに変換
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return nil, err
	}
	return graphJSON, nil
}
