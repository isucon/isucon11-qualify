package random

import (
	"sync"

	"github.com/docker/docker/pkg/namesgenerator"
)

var (
	mu            sync.Mutex
	generatedUser map[string]struct{}
)

func init() {
	generatedUser = make(map[string]struct{}, 128)
}

// 108 * 237 通りのユーザ名を重複なしで返す
func UserName() string {
	var username string
	retry := 0
	// NOTE: bench内から呼び出す処理で log.Fatalf して欲しくないので、無限ループする
	for {
		username = namesgenerator.GetRandomName(retry)
		if reserveName(username) {
			break
		}
		retry++
	}
	return username
}

//成功でtrue
func reserveName(username string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, exists := generatedUser[username]
	if exists {
		return false
	}
	generatedUser[username] = struct{}{}
	return true
}

func SetGeneratedUser(username string) {
	mu.Lock()
	defer mu.Unlock()
	generatedUser[username] = struct{}{}
}
