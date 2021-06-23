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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/service"
)

//汎用関数

func verifyStatusCode(hres *http.Response, code int) error {
	if hres.StatusCode != code {
		return errorInvalidStatusCode(hres)
	}
	return nil
}
func verifyContentType(hres *http.Response, contentType string) error {
	actual := hres.Header.Get("Content-Type")
	if !strings.HasPrefix(actual, contentType) {
		return errorInvalidContentType(hres, contentType)
	}
	return nil
}
func verifyJSONBody(hres *http.Response, body interface{}) error {
	decoder := json.NewDecoder(hres.Body)
	//defer hres.Body.Close()

	if err := decoder.Decode(body); err != nil {
		return errorInvalidJSON(hres)
	}
	return nil
}

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
	hres, err := a.Do(ctx, req)
	if err != nil {
		err = failure.NewError(ErrHTTP, err)
		errors = append(errors, err)
		return nil, errors
	}
	defer hres.Body.Close()

	//httpリクエストの検証
	res := &service.InitializeResponse{}
	err = verifyStatusCode(hres, http.StatusOK)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	//データの検証
	err = verifyContentType(hres, "application/json")
	if err != nil {
		errors = append(errors, err)
		//続行
	}
	err = verifyJSONBody(hres, &res)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}
	if res.Language == "" {
		err = errorBadResponse("利用言語(language)が設定されていません")
		errors = append(errors, err)
	}

	return res, errors
}
