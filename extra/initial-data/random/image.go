package random

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
)

func Image() []byte {
	files, err := ioutil.ReadDir(imageFolderPath)
	if err != nil {
		log.Fatalf("%+v", fmt.Errorf("%w", err))
	}
	fileInfo := files[rand.Intn(len(files))]

	bytes, err := ioutil.ReadFile(filepath.Join(imageFolderPath, fileInfo.Name()))
	if err != nil {
		log.Fatalf("%+v", fmt.Errorf("%w", err))
	}
	return bytes
}
