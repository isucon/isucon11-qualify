package model

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon11-qualify/extra/initial-data/random"
)

//enum
type UserType int

const (
	UserTypeNormal UserType = iota
	UserTypeMania
	UserTypeCompany
)

//基本的には一つのシナリオスレッドが一つのユーザーを占有する
//=>Isuの追加操作と、参照操作が同時に必要になる場面は無いはずなので、
//  IsuListのソートは追加が終わってからソートすれば良い
type User struct {
	UserID                  string `json:"jia_user_id"`
	Type                    UserType
	IsuListOrderByCreatedAt []*Isu          //CreatedAtは厳密にはわからないので、postした後にgetをした順番を正とする
	IsuListByID             map[string]*Isu //IDをkeyにアクセス
	Conditions              IsuConditionTreeSet

	Agent *agent.Agent
}

func NewRandomUserRaw(userType UserType) (*User, error) {
	return &User{
		UserID:                  random.UserName(),
		Type:                    userType,
		IsuListOrderByCreatedAt: []*Isu{},
		IsuListByID:             map[string]*Isu{},
		Agent:                   nil,
		Conditions:              NewIsuConditionTreeSet(),
	}, nil
}

//CreatedAt順で挿入すること
func (u *User) AddIsu(isu *Isu) {
	u.IsuListOrderByCreatedAt = append(u.IsuListOrderByCreatedAt, isu)
	u.IsuListByID[isu.JIAIsuUUID] = isu
}

func (user *User) GetConditionFromChan(ctx context.Context) {
	for _, isu := range user.IsuListOrderByCreatedAt {
		isu.getConditionFromChan(ctx, &user.Conditions)
	}
}
