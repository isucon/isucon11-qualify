package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
)

func (s *Scenario) InitializeData(ctx context.Context) {
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
		//var userConditions []model.IsuCondition

		for key, _ := range user.IsuListByID {
			// isu の初期化
			isu, _, err := model.NewIsuRawForInitData(user.IsuListByID[key], &user, key)
			if err != nil {
				logger.AdminLogger.Panicln(fmt.Errorf("初期データから User インスタンスを作成するのに失敗しました: %v", err))
			}
			//PosterForInitData(ctx, streamsForPoster)

			// isu.ID から model.TrendCondition を取得できるようにする (GET /trend 用)
			s.UpdateIsuFromID(isu)

			user.IsuListOrderByCreatedAt = append(user.IsuListOrderByCreatedAt, isu)
			//conditions := isu.GetIsuConditions()
		}
		sort.Slice(user.IsuListOrderByCreatedAt, func(i, j int) bool {
			return user.IsuListOrderByCreatedAt[i].CreatedAt.Before(user.IsuListOrderByCreatedAt[j].CreatedAt)
		})
		// sort.Slice(userConditions, func(i, j int) bool {
		// 	return userConditions[i].TimestampUnix < userConditions[j].TimestampUnix
		// })

		// user.Conditions = model.NewIsuConditionTreeSet()
		// for i, _ := range userConditions {
		// 	user.Conditions.Add(&userConditions[i])
		// }

		user.Type = model.UserTypeNormal
		s.normalUsers = append(s.normalUsers, &user)
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

// func PosterForInitData(ctx context.Context, stream *model.StreamsForPoster) {
// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 			}
// 			stream.ConditionChan <- []model.IsuCondition{}
// 		}
// 	}()
// }
