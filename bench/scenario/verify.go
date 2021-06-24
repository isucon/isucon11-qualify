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

// 文脈無しで検証できるもの
