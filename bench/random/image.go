package random

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/imgio"
)

const imageNum = 350

var index int32 = 0
var images [imageNum][]byte

func init() {
	var files []fs.FileInfo

	var err error
	// 画像ファイル群の読み込み
	files, err = ioutil.ReadDir(imageFolderPath)
	if err != nil {
		log.Fatalf("%+v", fmt.Errorf("%w", err))
	}

	for i := 0; i < imageNum; i++ {
		fileInfo := files[rand.Intn(len(files))]
		//default.jpg以外の、.jpgで終わるファイルに限定する
		for fileInfo.Name() == "default.jpg" || !strings.HasSuffix(fileInfo.Name(), ".jpg") {
			fileInfo = files[rand.Intn(len(files))]
		}
		img, err := imgio.Open(filepath.Join(imageFolderPath, fileInfo.Name()))
		if err != nil {
			log.Fatalf("%+v", err)
		}
		img = adjust.Brightness(img, float64(rand.Intn(20)-10)/10.0/2)
		img = adjust.Contrast(img, float64(rand.Intn(20)-10)/10.0/2)
		img = adjust.Gamma(img, 0.1+rand.Float64()*3)
		img = adjust.Saturation(img, float64(rand.Intn(20)-10)/10.0/2)

		//encode
		buffer := new(bytes.Buffer)
		encoder := imgio.JPEGEncoder(rand.Intn(95) + 5)
		encoder(buffer, img)
		images[i] = buffer.Bytes()
	}
}

func Image() ([]byte, error) {
	// MEMO: 現状 error は返してないがメモリがやばければファイル読み込みに変える
	return images[atomic.AddInt32(&index, 1)%imageNum], nil
}
