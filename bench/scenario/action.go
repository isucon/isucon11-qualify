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
	"net/http"

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
		err = errorBadResponse("利用言語(language)が設定されていません")
		errors = append(errors, err)
	}

	return initializeResponse, errors
}
