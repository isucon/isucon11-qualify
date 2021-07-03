package scenario

// verify.go
// 各種検証のユーティリティ関数
// ErrBadResponseのあたりの書式チェックと、
// シナリオのstructがあれば文脈無しで検証できるもの

import (
	"encoding/json"
	"net/http"
	"strings"
)

//汎用関数

func verifyStatusCode(res *http.Response, code int) error {
	if res.StatusCode != code {
		return errorInvalidStatusCode(res, code)
	}
	return nil
}
func verifyContentType(res *http.Response, contentType string) error {
	actual := res.Header.Get("Content-Type")
	if !strings.HasPrefix(actual, contentType) {
		return errorInvalidContentType(res, contentType)
	}
	return nil
}
func verifyJSONBody(res *http.Response, body interface{}) error {
	decoder := json.NewDecoder(res.Body)
	//defer res.Body.Close()

	if err := decoder.Decode(body); err != nil {
		return errorInvalidJSON(res)
	}
	return nil
}
func verifyText(res *http.Response, text string, expected string) error {
	if text != expected {
		return errorMissmatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", text, expected)
	}
	return nil
}

// 文脈無しで検証できるもの

func verifyNotSignedIn(res *http.Response, text string) error {
	if res.StatusCode != http.StatusUnauthorized || text != "you are not signed in" {
		// TODO: これ invalid status codeではないかも(textでの比較)
		return errorInvalidStatusCode(res, http.StatusUnauthorized)
	}
	return nil
}

// TODO: 統一され次第消す
func verifyNotSignedInTODO(res *http.Response, text string) error {
	if res.StatusCode != http.StatusUnauthorized || text != "you are not sign in" {
		// TODO: これ invalid status codeではないかも(textでの比較)
		return errorInvalidStatusCode(res, http.StatusUnauthorized)
	}
	return nil
}

func verifyBadReqBody(res *http.Response, text string) error {
	if res.StatusCode != http.StatusBadRequest || text != "bad request body" {
		// TODO: これ invalid status codeではないかも(textでの比較)
		return errorInvalidStatusCode(res, http.StatusBadRequest)
	}
	return nil
}

func verifyIsuNotFound(res *http.Response, text string) error {
	if res.StatusCode != http.StatusNotFound || text != "isu not found" {
		// TODO: これ invalid status codeではないかも(textでの比較)
		return errorInvalidStatusCode(res, http.StatusNotFound)
	}
	return nil
}
