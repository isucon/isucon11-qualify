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
	"strconv"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

func CheckError(err error) (critical bool, timeout bool, deduction bool) {
	critical = false  // クリティカルなエラー(起きたら即ベンチを止める)
	timeout = false   // リクエストタイムアウト(ある程度の数許容するかも)
	deduction = false // 減点対象になるエラー

	if isCritical(err) {
		critical = true
		return
	}

	if failure.IsCode(err, isucandar.ErrLoad) {
		if isTimeout(err) {
			timeout = true
		} else if isDeduction(err) {
			deduction = true
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
	return failure.IsCode(err, ErrCritical) ||
		failure.IsCode(err, ErrSecurityIncident) ||
		failure.IsCode(err, isucandar.ErrPanic)
}

var (
	ErrChecksum           failure.StringCode = "check-sum"
	ErrInvalidStatusCode  failure.StringCode = "status code"
	ErrInvalidContentType failure.StringCode = "content type"
	ErrInvalidJSON        failure.StringCode = "json"
	ErrInvalidAsset       failure.StringCode = "asset"
	ErrMismatch           failure.StringCode = "mismatch"     //データはあるが、間違っている（名前が違う等）
	ErrInvalid            failure.StringCode = "invalid"      //ロジック的に誤り（存在しないはずのものが有る等）
	ErrBadResponse        failure.StringCode = "bad-response" //不正な書式のレスポンス
	ErrHTTP               failure.StringCode = "http"         //http通信回りのエラー（timeout含む）
)

func isDeduction(err error) bool {
	return failure.IsCode(err, ErrInvalidStatusCode) ||
		failure.IsCode(err, ErrInvalidContentType) ||
		failure.IsCode(err, ErrInvalidJSON) ||
		failure.IsCode(err, ErrInvalidAsset) ||
		failure.IsCode(err, ErrMismatch) ||
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
	if failure.Is(err, context.DeadlineExceeded) ||
		failure.Is(err, context.Canceled) {
		return true
	}
	return failure.IsCode(err, failure.TimeoutErrorCode)
}

func IsValidation(err error) bool {
	return failure.IsCode(err, isucandar.ErrValidation)
}

func errorInvalidStatusCode(res *http.Response, expected int) error {
	return failure.NewError(ErrInvalidStatusCode, errorFormatWithResponse(res, "期待する HTTP ステータスコード以外が返却されました (expected: %d)", expected))
}

func errorInvalidStatusCodes(res *http.Response, expected []int) error {
	expectedStr := ""
	for _, v := range expected {
		expectedStr += strconv.Itoa(v) + ","
	}
	expectedStr = expectedStr[:len(expectedStr)-1]
	return failure.NewError(ErrInvalidStatusCode, errorFormatWithResponse(res, "期待する HTTP ステータスコード以外が返却されました (expected: %s)", expectedStr))
}

func errorInvalidContentType(res *http.Response, expected string) error {
	actual := res.Header.Get("Content-Type")
	return failure.NewError(ErrInvalidContentType,
		errorFormatWithResponse(res, "Content-Typeが正しくありません: %s (expected: %s)",
			actual, expected,
		))
}

func errorInvalidJSON(res *http.Response) error {
	return failure.NewError(ErrInvalidJSON, errorFormatWithResponse(res, "不正なJSONが返却されました"))
}

func errorMismatch(res *http.Response, message string, args ...interface{}) error {
	return failure.NewError(ErrMismatch, errorFormatWithResponse(res, message, args...))
}

func errorInvalid(res *http.Response, message string, args ...interface{}) error {
	return failure.NewError(ErrInvalid, errorFormatWithResponse(res, message, args...))
}

func errorCheckSum(message string, args ...interface{}) error {
	return failure.NewError(ErrChecksum, fmt.Errorf(message, args...))
}
func errorBadResponse(res *http.Response, message string, args ...interface{}) error {
	return failure.NewError(ErrBadResponse, errorFormatWithResponse(res, message, args...))
}

func errorFormatWithResponse(res *http.Response, message string, args ...interface{}) error {
	return errorFormatWithURI(res.StatusCode, res.Request.Method, res.Request.URL.RequestURI(), message, args...)
}
func errorFormatWithURI(statusCode int, method string, urlPath string, message string, args ...interface{}) error {
	args = append(args, statusCode, method, urlPath)
	return fmt.Errorf(message+": %d (%s: %s)", args...)
}
