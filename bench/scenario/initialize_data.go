package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"sync/atomic"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
)

func (s *Scenario) InitializeData(ctx context.Context) {
	raw, err := ioutil.ReadFile("./data/initialize.json")
	if err != nil {
		logger.AdminLogger.Panicln(fmt.Errorf("初期データファイルの読み込みに失敗しました: %v", err))
	}

	var users []model.User
	if err := json.Unmarshal(raw, &users); err != nil {
		logger.AdminLogger.Panicln(fmt.Errorf("初期データのParseに失敗しました: %v", err))
	}

	for i, _ := range users {
		user := &users[i]

		//var userConditions []model.IsuCondition

		for key, _ := range user.IsuListByID {
			// isu の初期化
			isu := user.IsuListByID[key]
			model.NewIsuRawForInitData(isu, user, key)
			//PosterForInitData(ctx, streamsForPoster)

			// isu.ID から model.TrendCondition を取得できるようにする (GET /trend 用)
			s.UpdateIsuFromID(isu)

			user.IsuListOrderByCreatedAt = append(user.IsuListOrderByCreatedAt, isu)
			//conditions := isu.GetIsuConditions()
		}
		sort.Slice(user.IsuListOrderByCreatedAt, func(i, j int) bool {
			return user.IsuListOrderByCreatedAt[i].ID < user.IsuListOrderByCreatedAt[j].ID
		})
		// sort.Slice(userConditions, func(i, j int) bool {
		// 	return userConditions[i].TimestampUnix < userConditions[j].TimestampUnix
		// })

		// user.Conditions = model.NewIsuConditionTreeSet()
		// for i, _ := range userConditions {
		// 	user.Conditions.Add(&userConditions[i])
		// }

		user.Type = model.UserTypeNormal
		s.normalUsers = append(s.normalUsers, user)
	}

	//初期データを登録
	for _, u := range s.normalUsers {
		random.SetGeneratedUser(u.UserID)
		atomic.StoreInt32(&u.PostIsuFinish, 1)
	}
}
