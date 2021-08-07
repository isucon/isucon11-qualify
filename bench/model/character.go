package model

import (
	"fmt"

	"github.com/isucon/isucon11-qualify/bench/random"
)

type IsuCharacterID int

func NewIsuCharacter(character string) (IsuCharacterID, error) {
	expected := random.CharacterData

	// contains check
	for id, v := range expected {
		if character == v {
			return IsuCharacterID(id), nil
		}
	}
	return -1, fmt.Errorf("性格が正しくありません")
}

type IsuCharacterSet []IsuCharacterID

// 重複チェック & append
func (cs IsuCharacterSet) Append(newCharacter IsuCharacterID) IsuCharacterSet {
	var flag bool
	for _, c := range cs {
		if c == newCharacter {
			flag = true
		}
	}
	if !flag {
		cs = append(cs, newCharacter)
	}
	return cs
}

// IsuCharacterSet の要素が全ての性格を持つのかの判定
func (cs IsuCharacterSet) IsFull() bool {
	// 要素名は IsuCharacter を new する時に確認済みなので、ここでは length のみを検証する
	return len(cs) == len(random.CharacterData)
}
