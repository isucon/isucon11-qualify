package model

import (
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon11-qualify/bench/random"
)

//enum
type UserType int

const (
	UserTypeNormal UserType = iota
)

//基本的には一つのシナリオ Goroutineが一つのユーザーを占有する
//=>Isuの追加操作と、参照操作が同時に必要になる場面は無いはずなので、
//  IsuListのソートは追加が終わってからソートすれば良い
type User struct {
	UserID                  string `json:"jia_user_id"`
	Type                    UserType
	IsuListOrderByCreatedAt []*Isu          //CreatedAtは厳密にはわからないので、並列postの場合はpostした後にgetをした順番を正とする
	IsuListByID             map[string]*Isu `json:"isu_list_by_id"` //IDをkeyにアクセス
	PostIsuFinish           int32

	Agent *agent.Agent
}

func NewRandomUserRaw(userType UserType, isIsuconUser bool) (*User, error) {
	var id string
	if isIsuconUser {
		id = "isucon"
	} else {
		id = random.UserName()
	}
	return &User{
		UserID:                  id,
		Type:                    userType,
		IsuListOrderByCreatedAt: []*Isu{},
		IsuListByID:             map[string]*Isu{},
		PostIsuFinish:           0,
		Agent:                   nil,
	}, nil
}

//CreatedAt順で挿入すること
func (u *User) AddIsu(isu *Isu) {
	u.IsuListOrderByCreatedAt = append(u.IsuListOrderByCreatedAt, isu)
	u.IsuListByID[isu.JIAIsuUUID] = isu
}

func (user *User) CloseAllIsuStateChan() {
	for _, isu := range user.IsuListByID {
		close(isu.StreamsForScenario.StateChan)
	}
}

func (u *User) GetAgent() *agent.Agent {
	return u.Agent
}
