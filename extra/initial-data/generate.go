package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/isucon/isucon11-qualify/extra/initial-data/models"
)

const (
	userNum           = 120
	userGeneralWeight = 2
	userManiaWeight   = 1
	userCompanyWeight = 3
)

func init() {
	t, _ := time.Parse(time.RFC3339, "2021-07-01T00:00:00+07:00")
	rand.Seed(t.UnixNano())
}

func main() {
	for i := 0; i < userNum; i++ {
		user := models.NewUser()
		if err := user.Create(); err != nil {
			log.Fatal(err)
		}

		// user の特性を乱数で決定
		var isuNum, durationMinute, conditionNum int
		n := rand.Intn(userGeneralWeight + userManiaWeight + userCompanyWeight)
		switch true {
		case n < userGeneralWeight: // 一般人
			// TODO: 定数やめる
			isuNum = 4
			durationMinute = 2
			conditionNum = 10
		case n < userGeneralWeight+userManiaWeight: // 一般人 (マニア)
			// TODO: 定数やめる
			isuNum = 20
			durationMinute = 1
			conditionNum = 10
		case n < userGeneralWeight+userManiaWeight+userCompanyWeight: // 企業
			// TODO: 定数やめる
			isuNum = 50
			durationMinute = 5
			conditionNum = 10
		}

		// User の所持する Isu 分だけ loop
		for j := 0; j < isuNum; j++ {
			isu := models.NewIsu(user)
			// 確率で Isu を更新
			if rand.Intn(4) < 1 { // 1/4
				isu = isu.WithUpdateName()
			}
			if rand.Intn(10) < 9 { // 9/10
				isu = isu.WithUpdateImage()
			}
			if rand.Intn(10) < 1 { // 1/10
				isu = isu.WithDelete()
			}
			// INSERT isu
			if err := isu.Create(); err != nil {
				log.Fatal(err)
			}

			// Isu の Condition 分だけ loop
			var condition models.Condition
			for k := 0; k < conditionNum; k++ {
				if k == 0 {
					condition = models.NewCondition(isu)
				} else {
					condition = models.NewConditionFromLastCondition(condition, durationMinute)
				}

				// INSERT condition & graph
				if err := condition.Create(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
