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
	rand.Seed(time.Now().UnixNano())
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

		for j := 0; j < isuNum; j++ {
			isu := models.NewIsu(user)
			if err := isu.Create(); err != nil {
				log.Fatal(err)
			}

			var condition models.Condition
			for k := 0; k < conditionNum; k++ {
				if k == 0 {
					condition = models.NewCondition(isu)
				} else {
					condition = models.NewConditionFromLastCondition(condition, durationMinute)
				}

				if k != conditionNum-1 {
					// INSERT condition
					if err := condition.Create(); err != nil {
						log.Fatal(err)
					}
				} else {
					// 最後の condition 挿入の場合、Graph を生成するために condition.Create の代わりに POST /api/isu/{jia_isu_uuid}/condition する
					if err := models.NewGraph(isu).CreateWithCondition(condition); err != nil {
						log.Fatal(err)
					}
				}
			}

		}
	}
}
