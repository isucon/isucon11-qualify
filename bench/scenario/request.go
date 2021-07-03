package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/isucon/isucandar/agent"
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
		return httpres, "", errorInvalidContentType(httpres, "text/plain")
	}

	return httpres, string(resBody), nil
}

func reqPngResNoContent(ctx context.Context, agent *agent.Agent, method string, rpath string, image io.Reader, allowedStatusCodes []int) (*http.Response, error) {
	body, contentType, err := getFormFromImage(image)
	if err != nil {
		return nil, err
	}
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", contentType)

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, err
	}

	return httpres, nil
}

func reqPngResError(ctx context.Context, agent *agent.Agent, method string, rpath string, image io.Reader, allowedStatusCodes []int) (*http.Response, string, error) {
	body, contentType, err := getFormFromImage(image)
	if err != nil {
		return nil, "", err
	}
	httpreq, err := agent.NewRequest(method, rpath, body)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpreq.Header.Set("Content-Type", contentType)

	httpres, err := doRequest(ctx, agent, httpreq, allowedStatusCodes)
	if err != nil {
		return nil, "", err
	}

	resBody, err := checkContentTypeAndGetBody(httpres, "text/plain")
	if err != nil {
		return httpres, "", errorInvalidContentType(httpres, "text/plain")
	}

	return httpres, string(resBody), nil
}

func getFormFromImage(image io.Reader) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part := make(textproto.MIMEHeader)
	part.Set("Content-Type", "image/png")
	part.Set("Content-Disposition", `form-data; name="image"; filename="image.png"`)
	pw, err := mw.CreatePart(part)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	_, err = io.Copy(pw, image)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	contentType := mw.FormDataContentType()
	err = mw.Close()
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	return body, contentType, nil
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

	// TODO: resBodyの扱いを考える(現状でここに置いてるのは Close 周りの都合)
	resBody, err := checkContentTypeAndGetBody(httpres, "image/png")
	if err != nil {
		return httpres, nil, errorInvalidContentType(httpres, "image/png")
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
		return httpres, errorInvalidContentType(httpres, "application/json")
	}

	if err := json.Unmarshal(resBody, res); err != nil {
		return nil, errorInvalidJSON(httpres)
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
		return httpres, "", errorInvalidContentType(httpres, "text/plain")
	}

	return httpres, string(resBody), nil
}

func doRequest(ctx context.Context, agent *agent.Agent, httpreq *http.Request, allowedStatusCodes []int) (*http.Response, error) {
	httpres, err := agent.Do(ctx, httpreq)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	invalidStatusCode := true
	for _, c := range allowedStatusCodes {
		if httpres.StatusCode == c {
			invalidStatusCode = false
		}
	}
	if invalidStatusCode {
		return nil, errorInvalidStatusCodes(httpres, allowedStatusCodes)
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
		return nil, err
	}

	return resBody, nil
}
