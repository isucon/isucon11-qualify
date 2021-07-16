package random

import (
	"log"
	"sync"

	"github.com/docker/docker/pkg/namesgenerator"
)

const (
	retry = 10
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
	for i := 0; i < retry; i++ { // 10 回連続で名前が被ったら exit 1 する
		username = namesgenerator.GetRandomName(0)
		if !hasAlreadyGenerated(username) {
			break
		}
		if i+1 == retry {
			log.Fatalf("username already exists (retry count: %d)", retry)
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
