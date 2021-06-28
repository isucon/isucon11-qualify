package scenario

// error.go
// エラー種類の定義
// エラー種類の判別関数
// エラーメッセージの構築補助関数

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

func CheckError(err error) (critical bool, timeout bool, deduction bool) {
	critical = false  // TODO: クリティカルなエラー(起きたら即ベンチを止める)
	timeout = false   // TODO: リクエストタイムアウト(ある程度の数許容するかも)
	deduction = false // TODO: 減点対象になるエラー

	if isCritical(err) {
		critical = true
		return
	}

	if failure.IsCode(err, isucandar.ErrLoad) {
		if isTimeout(err) {
			timeout = true
			return
		}
		if isDeduction(err) {
			timeout = true
		}
	}

	return
}

// Critical Errors
var (
	ErrCritical         failure.StringCode = "critical"
	ErrSecurityIncident failure.StringCode = "security incident"
)

func isCritical(err error) bool {
	// Prepare step でのエラーはすべて Critical の扱い
	return failure.IsCode(err, isucandar.ErrPrepare) ||
		failure.IsCode(err, ErrCritical) ||
		failure.IsCode(err, ErrSecurityIncident)
}

var (
	ErrInvalidStatusCode  failure.StringCode = "status code"
	ErrInvalidContentType failure.StringCode = "content type"
	ErrInvalidJSON        failure.StringCode = "json"
	ErrInvalidAsset       failure.StringCode = "asset"
	ErrMissmatch          failure.StringCode = "missmatch"    //データはあるが、間違っている（名前が違う等）
	ErrInvalid            failure.StringCode = "invalid"      //ロジック的に誤り（存在しないはずのものが有る等）
	ErrBadResponse        failure.StringCode = "bad-response" //不正な書式のレスポンス
	ErrHTTP               failure.StringCode = "http"         //http通信回りのエラー（timeout含む）
)

func isDeduction(err error) bool {
	return failure.IsCode(err, ErrInvalidStatusCode) ||
		failure.IsCode(err, ErrInvalidContentType) ||
		failure.IsCode(err, ErrInvalidJSON) ||
		failure.IsCode(err, ErrInvalidAsset) ||
		failure.IsCode(err, ErrMissmatch) ||
		failure.IsCode(err, ErrInvalid) ||
		failure.IsCode(err, ErrBadResponse) ||
		(!isTimeout(err) && failure.IsCode(err, ErrHTTP))
}

func isTimeout(err error) bool {
	var nerr net.Error
	if failure.As(err, &nerr) {
		if nerr.Timeout() || nerr.Temporary() {
			return true
		}
	}
	if failure.Is(err, context.DeadlineExceeded) {
		return true
	}
	return failure.IsCode(err, failure.TimeoutErrorCode)
}

func isValidation(err error) bool {
	return failure.IsCode(err, isucandar.ErrValidation)
}

func errorInvalidStatusCode(res *http.Response, expected int) error {
	return failure.NewError(ErrInvalidStatusCode, fmt.Errorf("期待する HTTP ステータスコード以外が返却されました (expected: %d): %d (%s: %s)", expected, res.StatusCode, res.Request.Method, res.Request.URL.Path))
}

func errorInvalidContentType(res *http.Response, expected string) error {
	actual := res.Header.Get("Content-Type")
	return failure.NewError(ErrInvalidContentType,
		fmt.Errorf("Content-Typeが正しくありません: %s (expected: %s): %d (%s: %s)",
			actual, expected, res.StatusCode, res.Request.Method, res.Request.URL.Path,
		))
}

func errorInvalidJSON(res *http.Response) error {
	return failure.NewError(ErrInvalidJSON, fmt.Errorf("不正なJSONが返却されました: %d (%s: %s)", res.StatusCode, res.Request.Method, res.Request.URL.Path))
}

func errorMissmatch(res *http.Response, message string, args ...interface{}) error {
	args = append(args, res.StatusCode, res.Request.Method, res.Request.URL.Path)
	return failure.NewError(ErrBadResponse, fmt.Errorf(message+": %d (%s: %s)", args...))
}

func errorBadResponse(res *http.Response, message string, args ...interface{}) error {
	args = append(args, res.StatusCode, res.Request.Method, res.Request.URL.Path)
	return failure.NewError(ErrBadResponse, fmt.Errorf(message+": %d (%s: %s)", args...))
}
