package controller

import (
	"crypto/ecdsa"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

/// Const Values ///

const (
	// lifetime は jwt の発行から失効までの期間を表す。
	lifetime = 30 * time.Minute
)

var (
	// 認証を通すユーザ/パスワード
	passwordMap = map[string]string{
		"isucon":  "isucon",
		"isucon1": "isucon1",
		"isucon2": "isucon2",
		"isucon3": "isucon3",
	}
)

/// Controller ///

type AuthController struct {
	jwtSecretKey *ecdsa.PrivateKey
}

func NewAuthController(key []byte) (*AuthController, error) {
	jwtSecretKey, err := jwt.ParseECPrivateKeyFromPEM(key)
	if err != nil {
		return nil, err
	}
	return &AuthController{jwtSecretKey}, nil
}

func (c *AuthController) PostAuth(ctx echo.Context) error {
	input := &struct {
		User     string `json:"user" validate:"required"`
		Password string `json:"password" validate:"required"`
	}{}
	err := ctx.Bind(input)
	if err != nil {
		ctx.Logger().Errorf("failed to bind: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	pass, ok := passwordMap[input.User]
	if !ok || pass != input.Password {
		return echo.NewHTTPError(http.StatusNotFound, "Not Found")
	}

	// 認証に利用する JWT トークンを生成して返す。
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"jia_user_id": input.User,
		"iat":         now.Unix(),
		"exp":         now.Add(lifetime).Unix(),
	})
	jwt, err := token.SignedString(c.jwtSecretKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return ctx.String(http.StatusOK, jwt)
}
