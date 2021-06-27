package service

import (
	"crypto/ecdsa"
	"io/ioutil"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecretKey *ecdsa.PrivateKey

func init() {
	jwtSecretKeyPath := "./key/ec256-private.pem"
	key, err := ioutil.ReadFile(jwtSecretKeyPath)
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}
	jwtSecretKey, err = jwt.ParseECPrivateKeyFromPEM(key)
	if err != nil {
		log.Fatalf("Unable to parse ECDSA private key: %v", err)
	}
}

// 認証に利用する JWT トークンを生成して返す。
func GenerateJWT(userID string, issuedAt time.Time) (string, error) {
	const lifetime = 30 * time.Second
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": userID,
		"iat":         issuedAt.Unix(),
		"exp":         issuedAt.Add(lifetime).Unix(),
	})

	return token.SignedString(jwtSecretKey)
}
