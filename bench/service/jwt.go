package service

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecretKey *ecdsa.PrivateKey

const lifetime = 30 * time.Second

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
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": userID,
		"iat":         issuedAt.Add(-1 * time.Second).Unix(), //#445 Token used before issued対策
		"exp":         issuedAt.Add(lifetime).Unix(),
	})

	return token.SignedString(jwtSecretKey)
}

// 異なる秘密鍵でJWTを生成する
func GenerateDummyJWT(userID string, issuedAt time.Time) (string, error) {
	jwtSecretDummyKeyPath := "./key/dummy.pem"
	key, err := ioutil.ReadFile(jwtSecretDummyKeyPath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %v", err)
	}
	jwtSecretDummyKey, err := jwt.ParseECPrivateKeyFromPEM(key)
	if err != nil {
		return "", fmt.Errorf("unable to parse ECDSA private key: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": userID,
		"iat":         issuedAt.Unix(),
		"exp":         issuedAt.Add(lifetime).Unix(),
	})

	return token.SignedString(jwtSecretDummyKey)
}

//異なる暗号方式でJWTを生成する
func GenerateHS256JWT(userID string, issuedAt time.Time) (string, error) {
	const secret = "wkZVBAb3DnUzvTPkDyD6WXNBTbqDNVqVrC0BuCPyey0l1ZFT1i6oag=="
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jia_user_id": userID,
		"iat":         issuedAt.Unix(),
		"exp":         issuedAt.Add(lifetime).Unix(),
	})

	return token.SignedString([]byte(secret))
}

//偽装したJWTを生成する
func GenerateTamperedJWT(userID1 string, userID2 string, issuedAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": userID1,
		"iat":         issuedAt.Unix(),
		"exp":         issuedAt.Add(lifetime).Unix(),
	})

	signed, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	//claimを置換する
	claims2Str := fmt.Sprintf(`{"jia_user_id":"%s","iat":%d,"exp":%d}`, userID2, issuedAt.Unix(), issuedAt.Add(lifetime).Unix())
	claims2 := jwt.EncodeSegment([]byte(claims2Str))
	jwtSep := strings.Split(signed, ".")
	return jwtSep[0] + "." + claims2 + "." + jwtSep[2], nil
}

//jia_user_idの無いJWTを生成する
func GenerateJWTWithNoData(issuedAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iat": issuedAt.Unix(),
		"exp": issuedAt.Add(lifetime).Unix(),
	})

	return token.SignedString(jwtSecretKey)
}

//jia_user_idの型がstringでないJWTを生成する
func GenerateJWTWithInvalidType(userID string, issuedAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": []interface{}{userID, issuedAt.Unix()},
		"iat":         issuedAt.Unix(),
		"exp":         issuedAt.Add(lifetime).Unix(),
	})

	return token.SignedString(jwtSecretKey)
}
