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

func UserName() string {
	var username string
	for { // bench内から呼び出す処理で log.Fatalf して欲しくないので、無限ループする
		username = namesgenerator.GetRandomName(0)
		if !hasAlreadyGenerated(username) {
			break
		}
	}
	setGeneratedUser(username)
	return username
}

func hasAlreadyGenerated(username string) bool {
	mu.Lock()
	defer mu.Unlock()
	_, exists := generatedUser[username]
	return exists
}

func setGeneratedUser(username string) {
	mu.Lock()
	defer mu.Unlock()
	generatedUser[username] = struct{}{}
}
