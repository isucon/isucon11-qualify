package random

import "math/rand"

func Character() string {
	return characterData[rand.Intn(len(characterData))]
}

var characterData = []string{
	"さみしがり",
	"いじっぱり",
	"やんちゃ",
	"ゆうかん",
	"ずぶとい",
	"わんぱく",
	"のうてんき",
	"のんき",
	"ひかえめ",
	"おっとり",
	"うっかりや",
	"れいせい",
	"おだやか",
	"おとなしい",
	"しんちょう",
	"なまいき",
	"おくびょう",
	"せっかち",
	"ようき",
	"むじゃき",
	"がんばりや",
	"すなお",
	"てれや",
	"きまぐれ",
	"まじめ",
}
