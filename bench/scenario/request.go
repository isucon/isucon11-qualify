package scenario

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"github.com/francoispqt/gojay"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/service"
	"github.com/pierrec/xxHash/xxHash64"
)

var (
	trendHash = TrendHash{mx: sync.Mutex{}, hash: map[uint64]service.GetTrendResponse{}}
	h64       = xxHash64.New(0)
)

type TrendHash struct {
	mx   sync.Mutex
	hash map[uint64]service.GetTrendResponse
}

func (t *TrendHash) getObj(res []byte) (service.GetTrendResponse, error) {
	t.mx.Lock()
	defer t.mx.Unlock()

	h64.Write(res)
	hash := h64.Sum64()
	h64.Reset()
	cache, exist := t.hash[hash]
	if exist {
		return cache, nil
	}

	obj := service.GetTrendResponse{}
	//err := gojay.UnmarshalJSONArray(res, &obj)
	err := json.Unmarshal(res, &obj)
	if err != nil {
		return nil, err
	}

	t.hash[hash] = obj
	return obj, nil
}

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

	if httpres.StatusCode != http.StatusNotModified {
		if !strings.HasPrefix(httpres.Header.Get("Content-Type"), "application/json") {
			return nil, errorInvalidContentType(httpres, "application/json")
		}
	}

	afterDec := gojay.NewDecoder(httpres.Body)
	defer afterDec.Release()
	err = afterDec.DecodeArray(res)
	if err != nil {
		return nil, err
	}

	return httpres, nil
}

func reqJSONResTrend(ctx context.Context, agent *agent.Agent, method string, rpath string, body io.Reader, allowedStatusCodes []int) (service.GetTrendResponse, *http.Response, error) {
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", "application/json")

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, nil, err
	}
	defer httpres.Body.Close()

	if httpres.StatusCode != http.StatusNotModified {
		if !strings.HasPrefix(httpres.Header.Get("Content-Type"), "application/json") {
			return nil, nil, errorInvalidContentType(httpres, "application/json")
		}
	}

	bytes, err := io.ReadAll(httpres.Body)
	if err != nil {
		return nil, nil, err
	}
	trend, err := trendHash.getObj(bytes)
	if err != nil {
		return nil, nil, err
	}

	return trend, httpres, nil
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
	httpres, err := AgentDo(agent, ctx, httpreq)
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

	if httpres.StatusCode != http.StatusNotModified {
		if !strings.HasPrefix(httpres.Header.Get("Content-Type"), contentType) {
			return nil, errorInvalidContentType(httpres, contentType)
		}
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
