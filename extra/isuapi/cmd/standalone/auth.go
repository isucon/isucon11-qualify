package main

import (
	"crypto/ecdsa"
	"io/ioutil"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	// iat と exp は登録済みクレーム名。それぞれの意味は https://tools.ietf.org/html/rfc7519#section-4.1 を参照。{
	iatKey = "iat"
	expKey = "exp"
	// }

	// lifetime は jwt の発行から失効までの期間を表す。
	lifetime = 30 * time.Minute
)

var jwtSecretKey *ecdsa.PrivateKey

func init() {
	jwtSecretKeyPath := getEnv("JWT_SECRETKEY_PATH", "./ec256-private.pem")
	key, err := ioutil.ReadFile(jwtSecretKeyPath)
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}
	jwtSecretKey, err = jwt.ParseECPrivateKeyFromPEM(key)
	if err != nil {
		log.Fatalf("Unable to parse ECDSA private key: %v", err)
	}
}

// Generate は認証に利用する JWT トークンを生成して返す。
func generateJWT(userID string, now time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": userID,
		"iat":         now.Unix(),
		"exp":         now.Add(lifetime).Unix(),
	})

	return token.SignedString(jwtSecretKey)
}
