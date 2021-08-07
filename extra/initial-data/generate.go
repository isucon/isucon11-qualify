package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/extra/initial-data/models"
)

const (
	userNum            = 320
	userPattern1Weight = 3
	userPattern2Weight = 1
	userPattern3Weight = 2
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
			// isucon ユーザは個人ユーザ相当（prepare checkに利用）
			{
				// isuconユーザは3日分のconditionを持つので、作成日は3日分マイナスしておく
				models.User{JIAUserID: "isucon", CreatedAt: random.Time().Add(-3 * 24 * time.Hour)},
				2,   // ISU の個数は 2
				10,  // condition を 10 分おきに送信
				432, // condition の総数は 72 時間分
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
				15,  // ISU の個数は 15
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
				isuCounter += 1
				isu := models.NewIsu(d.user)
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

	{ // insert data for random-generated user
		for i := 0; i < userNum; i++ {
			user := models.NewUser()
			if err := user.Create(); err != nil {
				log.Fatal(err)
			}
			isuListById := map[string]models.JsonIsuInfo{}

			// user の特性を乱数で決定
			var isuNum, conditionDurationMinutes, conditionNum int
			n := rand.Intn(userPattern1Weight + userPattern2Weight + userPattern3Weight)
			switch true {
			case n < userPattern1Weight: // 一般人
				isuNum = 1 + rand.Intn(15)        // isuは1-15で所有
				conditionDurationMinutes = 3      // condition を 3 分おきに送信
				conditionNum = 60 + rand.Intn(15) // condition 総数は 3 ~ 4 時間分
			case n < userPattern1Weight+userPattern2Weight: // 一般人 (コンディション多め)
				isuNum = 1 + rand.Intn(15)         // isuは1-15で所有
				conditionDurationMinutes = 5       // condition を 5 分おきに送信
				conditionNum = 100 + rand.Intn(25) // condition 総数は 8 ~ 10 時間分
			case n < userPattern1Weight+userPattern2Weight+userPattern3Weight: // 一般人（コンディションちょっと少なめ）
				isuNum = 1 + rand.Intn(15)        // isuは1-15で所有
				conditionDurationMinutes = 24     // condition を 24 分おきに送信
				conditionNum = 30 + rand.Intn(31) // condition 総数は 12 ~ 24 時間分
			}

			// User の所持する ISU 分だけ loop
			for j := 0; j < isuNum; j++ {
				isuCounter += 1
				isu := models.NewIsu(user)
				// 確率で ISU を更新
				if rand.Intn(4) < 1 { // 1/4
					if err := isu.WithUpdateName(); err != nil {
						log.Fatalf("%+v", err)
					}
				}
				if rand.Intn(10) < 9 { // 9/10
					if err := isu.WithUpdateImage(); err != nil {
						log.Fatalf("%+v", err)
					}
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
					if err := jsonConditions.AddCondition(condition, isuCounter); err != nil {
						log.Fatal(err)
					}
				}
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
