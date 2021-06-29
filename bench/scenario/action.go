package scenario

// action.go
// 1. リクエストを投げ
// 2. レスポンスを受け取り
// 3. ステータスコードを検証し
// 4. レスポンスをstructにマッピング
// 5. not nullのはずのフィールドのNULL値チェック
// までを行う
//
// 間で都度エラーチェックをする
// エラーが出た場合は返り値に入る

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/service"
)

//Action

func initializeAction(ctx context.Context, a *agent.Agent) (*service.InitializeResponse, []error) {
	errors := []error{}

	//リクエスト
	req, err := a.POST("/initialize", nil)
	if err != nil {
		err = failure.NewError(ErrHTTP, err)
		errors = append(errors, err)
		return nil, errors
	}
	res, err := a.Do(ctx, req)
	if err != nil {
		err = failure.NewError(ErrHTTP, err)
		errors = append(errors, err)
		return nil, errors
	}
	defer res.Body.Close()

	//httpリクエストの検証
	initializeResponse := &service.InitializeResponse{}
	err = verifyStatusCode(res, http.StatusOK)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	//データの検証
	err = verifyContentType(res, "application/json")
	if err != nil {
		errors = append(errors, err)
		//続行
	}
	err = verifyJSONBody(res, &initializeResponse)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}
	if initializeResponse.Language == "" {
		err = errorBadResponse(res, "利用言語(language)が設定されていません")
		errors = append(errors, err)
	}

	return initializeResponse, errors
}

func authAction(ctx context.Context, a *agent.Agent, userID string) (*service.AuthResponse, []error) {
	errors := []error{}

	//JWT生成
	jwtOK, err := service.GenerateJWT(userID, time.Now())
	if err != nil {
		err = failure.NewError(ErrCritical, err)
		errors = append(errors, err)
		return nil, errors
	}

	//リクエスト
	req, err := a.POST("/api/auth", nil)
	if err != nil {
		err = failure.NewError(ErrCritical, err)
		errors = append(errors, err)
		return nil, errors
	}
	req.Header.Set("Authorization", jwtOK)
	res, err := a.Do(ctx, req)
	if err != nil {
		err = failure.NewError(ErrHTTP, err)
		errors = append(errors, err)
		return nil, errors
	}
	defer res.Body.Close()

	//httpリクエストの検証
	authResponse := &service.AuthResponse{}
	err = verifyStatusCode(res, http.StatusOK)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	//データの検証
	//NoContentなので無し

	//Cookie検証
	found := false
	for _, c := range res.Cookies() {
		if c.Name == "isucondition" {
			found = true
			break
		}
	}
	if !found {
		err = errorBadResponse(res, "cookieがありません")
		errors = append(errors, err)
	}

	return authResponse, errors
}

func authActionWithInvalidJWT(ctx context.Context, a *agent.Agent, invalidJWT string, expectedCode int, expectedBody string) []error {
	errors := []error{}

	//リクエスト
	req, err := a.POST("/api/auth", nil)
	if err != nil {
		err = failure.NewError(ErrCritical, err)
		errors = append(errors, err)
		return errors
	}
	req.Header.Set("Authorization", invalidJWT)
	res, err := a.Do(ctx, req)
	if err != nil {
		err = failure.NewError(ErrHTTP, err)
		errors = append(errors, err)
		return errors
	}
	defer res.Body.Close()

	//httpリクエストの検証
	err = verifyStatusCode(res, expectedCode)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	//データの検証
	responseBody, _ := ioutil.ReadAll(res.Body)
	if string(responseBody) != expectedBody {
		err = errorMissmatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", string(responseBody), expectedBody)
		errors = append(errors, err)
	}

	return errors
}

//auth utlity

func authActionWithForbiddenJWT(ctx context.Context, a *agent.Agent, invalidJWT string) []error {
	return authActionWithInvalidJWT(ctx, a, invalidJWT, http.StatusForbidden, "forbidden")
}
