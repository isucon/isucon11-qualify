package random

import (
	"math/rand"
	"time"
)

const (
	imageFolderPath = "./images"
)

func init() {
	// generate random seed
	rand.Seed(time.Now().UnixNano())
}
