package scenario

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/isucon/isucon11-qualify/bench/model"
)

func (s *Scenario) InitializeData() {
	//TODO: Catalogは消す

	s.Catalogs = map[string]*model.IsuCatalog{
		"550e8400-e29b-41d4-a716-446655440000": {
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "isu0",
			LimitWeight: 150,
			Weight:      30,
			Size:        "W65.5×D66×H114.5~128.5cm",
			Maker:       "isu maker",
			Features:    "headrest,armrest",
		},
		"562dc0df-2d4f-4e38-98c0-9333f4ff3e38": {
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "isu1",
			LimitWeight: 136,
			Weight:      15,
			Size:        "W47×D43×H91cm～97cm",
			Maker:       "isu maker 2",
			Features:    "",
		},
	}

	raw, err := ioutil.ReadFile("./data/initial_data.json")
	if err != nil {
		panic(fmt.Errorf("初期データファイルの読み込みに失敗しました: %v", err))
	}
	// 一旦 Users に叩き込んでからnormalUsersとcompanyUsersに分離
	var users []model.User
	if err := json.Unmarshal(raw, &users); err != nil {
		panic(fmt.Errorf("初期データのParseに失敗しました: %v", err))
	}

	for i, _ := range users {
		user := users[i]

		for key, _ := range user.IsuListByID {
			isu := user.IsuListByID[key]
			isu.Owner = &user
			isu.JIAIsuUUID = key
			user.IsuListOrderByCreatedAt = append(user.IsuListOrderByCreatedAt, isu)
		}
		sort.Slice(user.IsuListOrderByCreatedAt, func(i, j int) bool {
			return user.IsuListOrderByCreatedAt[i].CreatedAt.Before(user.IsuListOrderByCreatedAt[j].CreatedAt)
		})

		switch len(user.IsuListByID) {
		case 4:
			user.Type = model.UserTypeNormal
			s.normalUsers = append(s.normalUsers, &user)
		case 20:
			user.Type = model.UserTypeMania
			s.normalUsers = append(s.normalUsers, &user)
		case 50:
			user.Type = model.UserTypeCompany
			s.companyUsers = append(s.companyUsers, &user)
		}
	}
}
