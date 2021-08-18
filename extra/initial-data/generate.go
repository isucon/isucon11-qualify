package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/extra/initial-data/models"
)

func init() {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	time.Local = loc
	t, _ := time.Parse(time.RFC3339, "2021-07-01T00:00:00+07:00")
	rand.Seed(t.UnixNano())
	uuid.SetRand(rand.New(rand.NewSource(t.UnixNano())))
}

func main() {
	jsonArray := models.JsonArray{}
	var isuCounter int
	{ // insert data for isucon user
		data := []struct {
			user                     models.User
			isuNum                   int
			conditionDurationMinutes int
			conditionNum             int
		}{
			// isucon ユーザは load 中に動かすため ISU は持たない
			// {
			// 	models.User{JIAUserID: "isucon", CreatedAt: random.Time().Add(-3 * 24 * time.Hour)},
			// },
			// isucon1 ユーザは個人ユーザ相当（prepare checkに利用）
			{
				// isucon1ユーザは3日分のconditionを持つので、作成日は3日分マイナスしておく
				models.User{JIAUserID: "isucon1", CreatedAt: random.Time().Add(-3 * 24 * time.Hour)},
				5,  // ISU の個数
				60, // condition を 60 分おきに送信
				72, // condition の総数は 72 時間分
			},
			// isucon2 ユーザは企業ユーザ相当
			{
				models.User{JIAUserID: "isucon2", CreatedAt: random.Time()},
				20, // ISU の個数
				30, // condition を 30 分おきに送信
				12, // condition の総数は 6 時間分
			},
			// isucon3 ユーザには isu を作成しない
			// {
			// 	models.User{JIAUserID: "isucon3", CreatedAt: random.Time()},
			// },
			// trend検証用のcondition生成用ユーザ
			{
				models.NewUser(),
				1,
				10,
				6,
			},
			{
				models.NewUser(),
				1,
				10,
				6,
			},
			{
				models.NewUser(),
				1,
				10,
				6,
			},
		}
		for _, d := range data {
			if err := d.user.Create(); err != nil {
				log.Fatal(err)
			}
			isuListById := map[string]models.JsonIsuInfo{}
			for j := 0; j < d.isuNum; j++ {
				isuCounter += 1
				// trendの検証のために全てのcharacterが必要なので25種類のcharacterを1つずつ用意できるようにしている
				characterId := isuCounter - 1
				// 同一character内のtimestampの順番検証のために同じ性格のISUを4つ確保する
				if isuCounter > 25 {
					characterId = 0
				}
				isu := models.NewIsuWithCharacterId(d.user, characterId)
				isu.CreatedAt = d.user.CreatedAt.Add(time.Minute) // ISU は User 作成の1分後に作成される
				if err := isu.WithUpdateName(); err != nil {
					log.Fatalf("%+v", err)
				}
				if err := isu.WithUpdateImage(); err != nil {
					log.Fatalf("%+v", err)
				}
				// INSERT isu
				if err := isu.Create(); err != nil {
					log.Fatal(err)
				}

				var jsonConditions models.JsonConditions
				// ISU の Condition 分だけ loop
				var condition models.Condition
				for k := 0; k < d.conditionNum; k++ {
					if k == 0 {
						condition = models.NewCondition(isu)
					} else {
						condition = models.NewConditionFromLastCondition(condition, d.conditionDurationMinutes)
					}
					// INSERT condition
					if err := condition.Create(); err != nil {
						log.Fatal(err)
					}
					// json用データ追加
					if err := jsonConditions.AddCondition(condition, isuCounter); err != nil {
						log.Fatal(err)
					}
				}
				isuListById[isu.JIAIsuUUID] = models.ToJsonIsuInfo(isuCounter, isu, jsonConditions)
			}
			jsonData := models.Json{
				JiaUserId:   d.user.JIAUserID,
				IsuListById: isuListById,
			}
			jsonArray = append(jsonArray, &jsonData)
		}
	}

	// ファイル出力
	if err := jsonArray.Commit(); err != nil {
		log.Fatal(err)
	}
}
