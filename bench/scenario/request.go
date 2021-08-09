package scenario

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/francoispqt/gojay"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
)

func reqNoContentResNoContent(ctx context.Context, agent *agent.Agent, method string, rpath string, allowedStatusCodes []int) (*http.Response, error) {
	httpreq, err := agent.NewRequest(method, rpath, nil)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, err
	}
	defer httpres.Body.Close()

	return httpres, nil
}

func reqNoContentResError(ctx context.Context, agent *agent.Agent, method string, rpath string, allowedStatusCodes []int) (*http.Response, string, error) {
	httpreq, err := agent.NewRequest(method, rpath, nil)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	// TODO: resBodyの扱いを考える(現状でここに置いてるのは Close 周りの都合)
	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, "", err
	}
	defer httpres.Body.Close()

	resBody, err := checkContentTypeAndGetBody(httpres, "text/plain")
	if err != nil {
		return httpres, "", err
	}

	return httpres, string(resBody), nil
}

func reqNoContentResPng(ctx context.Context, agent *agent.Agent, method string, rpath string, allowedStatusCodes []int) (*http.Response, []byte, error) {
	httpreq, err := agent.NewRequest(method, rpath, nil)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, nil, err
	}
	defer httpres.Body.Close()

	//ContentTypeのチェックは行わない
	//resBody, err := checkContentTypeAndGetBody(httpres, "image/png")
	resBody, err := ioutil.ReadAll(httpres.Body)
	if err != nil {
		// if !isTimeout(err) {
		// 	return httpres, nil, failure.NewError(ErrCritical, err)
		// }

		//MEMO: 仕様をよく知らず、想定外のエラーを全部Criticalにするのが怖いので逃げておく by eiya
		return httpres, nil, failure.NewError(ErrHTTP, err)
	}

	return httpres, resBody, nil
}

func reqJSONResJSON(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, res interface{}, allowedStatusCodes []int) (*http.Response, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", "application/json")

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, err
	}
	defer httpres.Body.Close()

	resBody, err := checkContentTypeAndGetBody(httpres, "application/json")
	if err != nil {
		return httpres, err
	}

	if err := json.Unmarshal(resBody, res); err != nil {
		return nil, errorInvalidJSON(httpres)
	}

	return httpres, nil
}

func reqJSONResGojayArray(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, res gojay.UnmarshalerJSONArray, allowedStatusCodes []int) (*http.Response, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", "application/json")

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, err
	}
	defer httpres.Body.Close()

	if !strings.HasPrefix(httpres.Header.Get("Content-Type"), "application/json") {
		return nil, errorInvalidContentType(httpres, "application/json")
	}

	dec := gojay.NewDecoder(httpres.Body)
	defer dec.Release()
	err = dec.DecodeArray(res)
	if err != nil {
		return nil, err
	}

	return httpres, nil
}

func reqJSONResNoContent(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, allowedStatusCodes []int) (*http.Response, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", "application/json")

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, err
	}
	defer httpres.Body.Close()

	return httpres, nil
}

func reqJSONResError(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, allowedStatusCodes []int) (*http.Response, string, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", "application/json")

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, "", err
	}

	resBody, err := checkContentTypeAndGetBody(httpres, "text/plain")
	if err != nil {
		return httpres, "", err
	}

	return httpres, string(resBody), nil
}

func reqMultipartResJSON(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, writer *multipart.Writer, res interface{}, allowedStatusCodes []int) (*http.Response, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", writer.FormDataContentType())

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return httpres, err
	}
	defer httpres.Body.Close()

	resBody, err := checkContentTypeAndGetBody(httpres, "application/json")
	if err != nil {
		return httpres, err
	}

	if err := json.Unmarshal(resBody, res); err != nil {
		return nil, errorInvalidJSON(httpres)
	}

	return httpres, nil
}

func reqMultipartResError(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, writer *multipart.Writer, allowedStatusCodes []int) (*http.Response, string, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", writer.FormDataContentType())

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, "", err
	}

	resBody, err := checkContentTypeAndGetBody(httpres, "text/plain")
	if err != nil {
		return httpres, "", err
	}

	return httpres, string(resBody), nil
}

func doRequest(ctx context.Context, agent *agent.Agent, httpreq *http.Request, allowedStatusCodes []int) (*http.Response, error) {
	httpres, err := agent.Do(ctx, httpreq)
	if err != nil {
		return nil, failure.NewError(ErrHTTP, err)
	}

	invalidStatusCode := true
	if httpreq.Method == http.MethodGet {
		allowedStatusCodes = append(allowedStatusCodes, http.StatusNotModified)
	}
	for _, c := range allowedStatusCodes {
		if httpres.StatusCode == c {
			invalidStatusCode = false
		}
	}
	if invalidStatusCode {
		return httpres, errorInvalidStatusCodes(httpres, allowedStatusCodes)
	}

	return httpres, nil
}

func checkContentTypeAndGetBody(httpres *http.Response, contentType string) ([]byte, error) {
	defer httpres.Body.Close()

	if !strings.HasPrefix(httpres.Header.Get("Content-Type"), contentType) {
		return nil, errorInvalidContentType(httpres, contentType)
	}

	resBody, err := ioutil.ReadAll(httpres.Body)
	if err != nil {
		// if !isTimeout(err) {
		// 	return nil, failure.NewError(ErrCritical, err)
		// }

		//MEMO: 仕様をよく知らず、想定外のエラーを全部Criticalにするのが怖いので逃げておく by eiya
		return nil, failure.NewError(ErrHTTP, err)
	}

	return resBody, nil
}
