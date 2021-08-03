package scenario

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
)

func (s *Scenario) InitializeData() {
	//TODO: ちゃんと生成する

	raw, err := ioutil.ReadFile("./data/initialize.json")
	if err != nil {
		logger.AdminLogger.Panicln(fmt.Errorf("初期データファイルの読み込みに失敗しました: %v", err))
	}

	var users []model.User
	if err := json.Unmarshal(raw, &users); err != nil {
		logger.AdminLogger.Panicln(fmt.Errorf("初期データのParseに失敗しました: %v", err))
	}

	for i, _ := range users {
		user := users[i]
		var userConditions []model.IsuCondition

		for key, _ := range user.IsuListByID {
			isu := user.IsuListByID[key]
			isu.Owner = &user
			isu.JIAIsuUUID = key
			user.IsuListOrderByCreatedAt = append(user.IsuListOrderByCreatedAt, isu)
			for _, cond := range isu.Conditions.Info {
				userConditions = append(userConditions, cond)
			}
			for _, cond := range isu.Conditions.Warning {
				userConditions = append(userConditions, cond)
			}
			for _, cond := range isu.Conditions.Critical {
				userConditions = append(userConditions, cond)
			}
		}
		sort.Slice(user.IsuListOrderByCreatedAt, func(i, j int) bool {
			return user.IsuListOrderByCreatedAt[i].ID < user.IsuListOrderByCreatedAt[j].ID
		})
		sort.Slice(userConditions, func(i, j int) bool {
			return userConditions[i].TimestampUnix < userConditions[j].TimestampUnix
		})

		user.Conditions = model.NewIsuConditionTreeSet()
		for i, _ := range userConditions {
			user.Conditions.Add(&userConditions[i])
		}

		user.Type = model.UserTypeNormal
		s.normalUsers = append(s.normalUsers, &user)
	}

	//初期データを登録
	for _, u := range s.normalUsers {
		random.SetGeneratedUser(u.UserID)
	}

	//for debug
	//{
	//	for _, user := range s.normalUsers {
	//		logger.AdminLogger.Printf("user: %#v\n", user)
	//		logger.AdminLogger.Printf("user info conds: %#v\n", user.Conditions.Info)
	//		logger.AdminLogger.Printf("user warn conds: %#v\n", user.Conditions.Warning)
	//		logger.AdminLogger.Printf("user crit conds: %#v\n", user.Conditions.Critical)
	//
	//		for id, isu := range user.IsuListByID {
	//			logger.AdminLogger.Printf("isu_id: %#v\n", id)
	//			logger.AdminLogger.Printf("isu_info: %v %v %v %v %v\n", isu.ID, isu.Name, fmt.Sprintf("%x", isu.ImageHash), isu.Character, isu.CreatedAt)
	//			logger.AdminLogger.Printf("info len: %#v\n", len(isu.Conditions.Info))
	//			for _, cond := range isu.Conditions.Info {
	//				logger.AdminLogger.Printf("cond: %#v\n", cond)
	//				break
	//			}
	//			logger.AdminLogger.Printf("warn len: %#v\n", len(isu.Conditions.Warning))
	//			for _, cond := range isu.Conditions.Warning {
	//				logger.AdminLogger.Printf("cond: %#v\n", cond)
	//				break
	//			}
	//			logger.AdminLogger.Printf("critical len: %#v\n", len(isu.Conditions.Critical))
	//			for _, cond := range isu.Conditions.Critical {
	//				logger.AdminLogger.Printf("cond: %#v\n", cond)
	//				break
	//			}
	//		}
	//		break
	//	}
	//	logger.AdminLogger.Printf("normal users len: %v", len(s.normalUsers))
	//}
}
