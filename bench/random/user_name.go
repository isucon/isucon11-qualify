package random

import (
	"log"
	"sync"

	"github.com/docker/docker/pkg/namesgenerator"
)

const (
	retry = 10000
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
	for i := 0; true; i++ { // bench内から呼び出す処理で log.Fatalf して欲しくないので、無限ループする
		username = namesgenerator.GetRandomName(0)
		if !hasAlreadyGenerated(username) {
			break
		}
		if i == retry-1 {
			log.Printf("[WARNING] username generating is probably in an infinite loop: already retried %d times\n", retry)
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
