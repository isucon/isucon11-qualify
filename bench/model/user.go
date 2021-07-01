package model

import (
	"math/rand"
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
	//ここで[]IsuLogを持つと更新にmutexが必要で嫌なので持たない
}

func (u *User) AddIsu(isu *Isu) {
	//TODO
}

// utility

func MakeRandomUserID() (string, error) {
	//TODO: とりあえず完全乱数だけど、ちゃんとそれっぽいのを生成しても良いかも
	//TODO: 重複削除

	const digit = 10
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, digit)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// letters からランダムに取り出して文字列を生成
	for i, v := range b {
		// index が letters の長さに収まるように調整
		b[i] = letters[int(v)%len(letters)]
	}
	return string(b), nil
}
