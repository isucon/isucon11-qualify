package random

import (
	"github.com/docker/docker/pkg/namesgenerator"
)

func UserName() string {
	// MEMO: すでに存在するユーザ名を出力する可能性がある
	return namesgenerator.GetRandomName(0)
}
