package model

import (
	"math/rand"
)

//TODO: Userのstructを書く

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
