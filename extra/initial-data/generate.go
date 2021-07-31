package main

import (
	"github.com/google/uuid"
	"log"
	"math/rand"
	"time"

	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/extra/initial-data/models"
)

const (
	userNum           = 120
	userGeneralWeight = 3
	userManiaWeight   = 1
	userCompanyWeight = 2
)

func init() {
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
			// isucon ユーザは個人ユーザ相当
			{
				models.User{JIAUserID: "isucon", CreatedAt: random.Time()},
				2,   // ISU の個数は 2
				3,   // condition を 3 分おきに送信
				480, // condition の総数は 24 時間分
			},
			// isucon1 ユーザは個人ユーザ相当
			{
				models.User{JIAUserID: "isucon1", CreatedAt: random.Time()},
				2,   // ISU の個数は 2
				3,   // condition を 3 分おきに送信
				480, // condition の総数は 24 時間分
			},
			// isucon2 ユーザは企業ユーザ相当
			{
				models.User{JIAUserID: "isucon2", CreatedAt: random.Time()},
				50,  // ISU の個数は 50
				5,   // condition を 5 分おきに送信
				288, // condition の総数は 24 時間分
			},
			// isucon3 ユーザには isu を作成しない
			// {
			// 	models.User{JIAUserID: "isucon3", CreatedAt: random.Time()},
			// },
		}
		for _, d := range data {
			if err := d.user.Create(); err != nil {
				log.Fatal(err)
			}
			isuListById := map[string]models.JsonIsuInfo{}
			for j := 0; j < d.isuNum; j++ {
				isu := models.NewIsu(d.user)
				isu.CreatedAt = d.user.CreatedAt.Add(time.Minute) // ISU は User 作成の1分後に作成される
				isu = isu.WithUpdateName()
				isu = isu.WithUpdateImage()
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
					if err := jsonConditions.AddCondition(condition); err != nil {
						log.Fatal(err)
					}
				}
				isuCounter += 1
				isuListById[isu.JIAIsuUUID] = models.ToJsonIsuInfo(isuCounter, isu, jsonConditions)
			}
			jsonData := models.Json{
				JiaUserId:   d.user.JIAUserID,
				IsuListById: isuListById,
			}
			jsonArray = append(jsonArray, &jsonData)
		}
	}

	{ // insert data for random-generated user
		for i := 0; i < userNum; i++ {
			user := models.NewUser()
			if err := user.Create(); err != nil {
				log.Fatal(err)
			}
			isuListById := map[string]models.JsonIsuInfo{}

			// user の特性を乱数で決定
			var isuNum, conditionDurationMinutes, conditionNum int
			n := rand.Intn(userGeneralWeight + userManiaWeight + userCompanyWeight)
			switch true {
			case n < userGeneralWeight: // 一般人
				isuNum = 1 + rand.Intn(4)         // ISU の個数は 1 ~ 4
				conditionDurationMinutes = 3      // condition を 3 分おきに送信
				conditionNum = 60 + rand.Intn(20) // condition 総数は 3 ~ 4 時間分
			case n < userGeneralWeight+userManiaWeight: // 一般人 (マニア)
				isuNum = 15 + rand.Intn(10)        // ISU の個数は 15 ~ 24
				conditionDurationMinutes = 1       // condition を 1 分おきに送信
				conditionNum = 180 + rand.Intn(60) // condition 総数は 3 ~ 4 時間分
			case n < userGeneralWeight+userManiaWeight+userCompanyWeight: // 企業
				isuNum = 40 + rand.Intn(20)       // ISU の個数は 40 ~ 59
				conditionDurationMinutes = 5      // condition を 5 分おきに送信
				conditionNum = 36 + rand.Intn(12) // condition 総数は 3 ~ 4 時間分
			}

			// User の所持する ISU 分だけ loop
			for j := 0; j < isuNum; j++ {
				isu := models.NewIsu(user)
				// 確率で ISU を更新
				if rand.Intn(4) < 1 { // 1/4
					isu = isu.WithUpdateName()
				}
				if rand.Intn(10) < 9 { // 9/10
					isu = isu.WithUpdateImage()
				}
				// INSERT isu
				if err := isu.Create(); err != nil {
					log.Fatal(err)
				}

				var jsonConditions models.JsonConditions
				// ISU の Condition 分だけ loop
				var condition models.Condition
				for k := 0; k < conditionNum; k++ {
					if k == 0 {
						condition = models.NewCondition(isu)
					} else {
						condition = models.NewConditionFromLastCondition(condition, conditionDurationMinutes)
					}

					// INSERT condition
					if err := condition.Create(); err != nil {
						log.Fatal(err)
					}
					// json用データ追加
					if err := jsonConditions.AddCondition(condition); err != nil {
						log.Fatal(err)
					}
				}
				isuCounter += 1
				isuListById[isu.JIAIsuUUID] = models.ToJsonIsuInfo(isuCounter, isu, jsonConditions)
			}
			jsonData := models.Json{
				JiaUserId:   user.JIAUserID,
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
