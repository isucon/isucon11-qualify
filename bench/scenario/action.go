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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

const (
	homeIsuLimit   = 4
	conditionLimit = 20
	isuListLimit   = 200 // TODO 修正が必要なら変更
)

//Action

// ==============================initialize==============================

func initializeAction(ctx context.Context, a *agent.Agent, req service.PostInitializeRequest) (*service.InitializeResponse, []error) {
	errors := []error{}

	//リクエスト
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	initializeResponse := &service.InitializeResponse{}
	res, err := reqJSONResJSON(ctx, a, http.MethodPost, "/initialize", bytes.NewReader(body), &initializeResponse, []int{http.StatusOK})
	if err != nil {
		errors = append(errors, err)
	} else {
		//データの検証
		if initializeResponse.Language == "" {
			err = errorBadResponse(res, "利用言語(language)が設定されていません")
			errors = append(errors, err)
		}
	}

	return initializeResponse, errors
}

// ==============================authAction==============================

func authAction(ctx context.Context, a *agent.Agent, userID string) (*service.AuthResponse, []error) {
	errors := []error{}

	//JWT生成
	jwtOK, err := service.GenerateJWT(userID, time.Now())
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	//リクエスト
	req, err := a.POST("/api/auth", nil)
	if err != nil {
		logger.AdminLogger.Panic(err)
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
		logger.AdminLogger.Panic(err)
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

func authActionWithoutJWT(ctx context.Context, a *agent.Agent) []error {
	errors := []error{}

	//リクエスト
	req, err := a.POST("/api/auth", nil)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	res, err := a.Do(ctx, req)
	if err != nil {
		err = failure.NewError(ErrHTTP, err)
		errors = append(errors, err)
		return errors
	}
	defer res.Body.Close()

	//httpリクエストの検証
	err = verifyStatusCode(res, http.StatusForbidden)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	//データの検証
	const expectedBody = "forbidden"
	responseBody, _ := ioutil.ReadAll(res.Body)
	if string(responseBody) != expectedBody {
		err = errorMissmatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", string(responseBody), expectedBody)
		errors = append(errors, err)
	}

	return errors
}

//auth utility

const authActionErrorNum = 8 //authActionErrorが何種類のエラーを持っているか

//正しく失敗するか確認するAction
func authActionError(ctx context.Context, agt *agent.Agent, userID string, errorType int) []error {
	switch errorType % authActionErrorNum {
	case 0:
		//Unexpected signing method, StatusForbidden
		jwtHS256, err := service.GenerateHS256JWT(userID, time.Now())
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		return authActionWithForbiddenJWT(ctx, agt, jwtHS256)
	case 1:
		//expired, StatusForbidden
		jwtExpired, err := service.GenerateJWT(userID, time.Now().Add(-365*24*time.Hour))
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		return authActionWithForbiddenJWT(ctx, agt, jwtExpired)
	case 2:
		//jwt is missing, StatusForbidden
		return authActionWithoutJWT(ctx, agt)
	case 3:
		//invalid private key, StatusForbidden
		jwtDummyKey, err := service.GenerateDummyJWT(userID, time.Now())
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		return authActionWithForbiddenJWT(ctx, agt, jwtDummyKey)
	case 4:
		//not jwt, StatusForbidden
		return authActionWithForbiddenJWT(ctx, agt, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.")
	case 5:
		//偽装されたjwt, StatusForbidden
		userID2 := random.UserName()
		jwtTampered, err := service.GenerateTamperedJWT(userID, userID2, time.Now())
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		return authActionWithForbiddenJWT(ctx, agt, jwtTampered)
	case 6:
		//jia_user_id is missing, StatusBadRequest
		jwtNoData, err := service.GenerateJWTWithNoData(time.Now())
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		return authActionWithInvalidJWT(ctx, agt, jwtNoData, http.StatusBadRequest, "invalid JWT payload")
	case 7:
		//jwt with invalid data type, StatusBadRequest
		jwtInvalidDataType, err := service.GenerateJWTWithInvalidType(userID, time.Now())
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		return authActionWithInvalidJWT(ctx, agt, jwtInvalidDataType, http.StatusBadRequest, "invalid JWT payload")
	}

	//ロジック的に到達しないはずだけど念のためエラー処理
	err := fmt.Errorf("internal bench error @authActionError: errorType=%v", errorType)
	logger.AdminLogger.Panic(err)
	return []error{failure.NewError(ErrCritical, err)}
}

func authActionWithForbiddenJWT(ctx context.Context, a *agent.Agent, invalidJWT string) []error {
	return authActionWithInvalidJWT(ctx, a, invalidJWT, http.StatusForbidden, "forbidden")
}

func signoutAction(ctx context.Context, a *agent.Agent) (*http.Response, error) {
	res, err := reqNoContentResNoContent(ctx, a, http.MethodPost, "/api/signout", []int{http.StatusOK})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func signoutErrorAction(ctx context.Context, a *agent.Agent) (string, *http.Response, error) {
	res, resBody, err := reqNoContentResError(ctx, a, http.MethodPost, "/api/signout", []int{http.StatusUnauthorized})
	if err != nil {
		return resBody, nil, err
	}
	return resBody, res, nil
}

func getMeAction(ctx context.Context, a *agent.Agent) (*service.GetMeResponse, *http.Response, error) {
	me := &service.GetMeResponse{}
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, "/api/user/me", nil, me, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return me, res, nil
}

func getMeErrorAction(ctx context.Context, a *agent.Agent) (string, *http.Response, error) {
	res, resBody, err := reqJSONResError(ctx, a, http.MethodGet, "/api/user/me", nil, []int{http.StatusUnauthorized})
	if err != nil {
		return resBody, nil, err
	}
	return resBody, res, nil
}

func getIsuAction(ctx context.Context, a *agent.Agent) ([]*service.Isu, *http.Response, error) {
	targetURL, err := url.Parse("/api/isu")
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}

	var isuList []*service.Isu
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, targetURL.String(), nil, &isuList, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return isuList, res, nil
}

func getIsuErrorAction(ctx context.Context, a *agent.Agent) (string, *http.Response, error) {
	targetURL, err := url.Parse("/api/isu")
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}

	res, resBody, err := reqJSONResError(ctx, a, http.MethodGet, targetURL.String(), nil, []int{http.StatusUnauthorized, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return resBody, res, nil
}

func getPathWithParams(pathStr string, query url.Values) string {
	path, err := url.Parse(pathStr)
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}

	path.RawQuery = query.Encode()
	return path.String()
}

func postIsuAction(ctx context.Context, a *agent.Agent, req service.PostIsuRequest) (*service.Isu, *http.Response, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormField("jia_isu_uuid")
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	_, err = part.Write([]byte(req.JIAIsuUUID))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	part, err = writer.CreateFormField("isu_name")
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	_, err = part.Write([]byte(req.IsuName))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	if req.Img != nil {
		partHeader := textproto.MIMEHeader{}
		partHeader.Set("Content-Type", "image/jpeg")
		partHeader.Set("Content-Disposition", `form-data; name="image"; filename="image.jpeg"`)

		part, err := writer.CreatePart(partHeader)
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
		_, err = part.Write(req.Img)
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
	}

	err = writer.Close()
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	isu := &service.Isu{}
	res, err := reqMultipartResJSON(ctx, a, http.MethodPost, "/api/isu", buf, writer, isu, []int{http.StatusCreated})
	if err != nil {
		return nil, res, err
	}
	return isu, res, nil
}

func postIsuErrorAction(ctx context.Context, a *agent.Agent, req service.PostIsuRequest) (string, *http.Response, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormField("jia_isu_uuid")
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	_, err = part.Write([]byte(req.JIAIsuUUID))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	part, err = writer.CreateFormField("isu_name")
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	_, err = part.Write([]byte(req.IsuName))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	// TODO: 画像も追加する
	// TODO: file.Nameを正しく渡すと不正出来そうなので、拡張子残すくらいにしておきたい
	//part, err := writer.CreateFormFile("image", filepath.Base(file.Name()))
	//io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	res, text, err := reqMultipartResError(ctx, a, http.MethodPost, "/api/isu", buf, writer, []int{http.StatusBadRequest, http.StatusConflict, http.StatusUnauthorized, http.StatusNotFound, http.StatusForbidden})
	if err != nil {
		return "", res, err
	}
	return text, res, nil
}

func getIsuIdAction(ctx context.Context, a *agent.Agent, id string) (*service.Isu, *http.Response, error) {
	isu := &service.Isu{}
	reqUrl := fmt.Sprintf("/api/isu/%s", id)
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &isu, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return isu, res, nil
}

func getIsuIdErrorAction(ctx context.Context, a *agent.Agent, id string) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s", id)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusUnauthorized, http.StatusNotFound})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getIsuIconAction(ctx context.Context, a *agent.Agent, id string, allowNotModified bool) ([]byte, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/icon", id)
	allowedStatusCodes := []int{http.StatusOK}
	if allowNotModified {
		allowedStatusCodes = append(allowedStatusCodes, http.StatusNotModified)
	}
	res, image, err := reqNoContentResPng(ctx, a, http.MethodGet, reqUrl, allowedStatusCodes)
	if err != nil {
		return nil, nil, err
	}
	// TODO: imageの取り扱いについて考える
	return image, res, nil
}

func getIsuIconErrorAction(ctx context.Context, a *agent.Agent, id string) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/icon", id)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusUnauthorized, http.StatusNotFound})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func postIsuConditionAction(ctx context.Context, httpClient http.Client, targetUrl string, req *[]service.PostIsuConditionRequest) (*http.Response, error) {
	conditionByte, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", targetUrl, bytes.NewBuffer(conditionByte))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "JIA-Members-Client/1.2")
	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := verifyStatusCodes(res, []int{http.StatusCreated, http.StatusServiceUnavailable}); err != nil {
		return nil, err
	}
	return res, nil
}

func postIsuConditionErrorAction(ctx context.Context, httpClient http.Client, targetUrl string, req []map[string]interface{}) (string, *http.Response, error) {
	conditionByte, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", targetUrl, bytes.NewBuffer(conditionByte))
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "JIA-Members-Client/1.2")
	res, err := httpClient.Do(httpReq)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()

	resBody, err := checkContentTypeAndGetBody(res, "text/plain")
	if err != nil {
		return "", nil, err
	}

	return string(resBody), res, nil
}

func getIsuConditionAction(ctx context.Context, a *agent.Agent, id string, req service.GetIsuConditionRequest) ([]*service.GetIsuConditionResponse, *http.Response, error) {
	reqUrl := getIsuConditionRequestParams(fmt.Sprintf("/api/condition/%s", id), req)
	conditions := []*service.GetIsuConditionResponse{}
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &conditions, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return conditions, res, nil
}

func getIsuConditionErrorAction(ctx context.Context, a *agent.Agent, id string, query url.Values) (string, *http.Response, error) {
	path := fmt.Sprintf("/api/condition/%s", id)
	rpath := getPathWithParams(path, query)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, rpath, []int{http.StatusNotFound, http.StatusUnauthorized, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getIsuConditionRequestParams(base string, req service.GetIsuConditionRequest) string {
	targetURL, err := url.Parse(base)
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}

	q := url.Values{}
	if req.StartTime != nil {
		q.Set("start_time", fmt.Sprint(*req.StartTime))
	}
	q.Set("end_time", fmt.Sprint(req.EndTime))
	q.Set("condition_level", req.ConditionLevel)
	if req.Limit != nil {
		q.Set("limit", fmt.Sprint(*req.Limit))
	}
	targetURL.RawQuery = q.Encode()
	return targetURL.String()
}

func getIsuGraphAction(ctx context.Context, a *agent.Agent, id string, req service.GetGraphRequest) (service.GraphResponse, *http.Response, error) {
	graph := service.GraphResponse{}
	reqUrl := fmt.Sprintf("/api/isu/%s/graph?datetime=%d", id, req.Date)
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &graph, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}

	//TODO: バリデーション
	// res, text, err := reqJSONResError(ctx, a, http.MethodPost, reqUrl, bytes.NewReader(body), []int{http.StatusNotFound, http.StatusBadRequest})
	// if err != nil {
	// 	return "", nil, err
	// }

	return graph, res, nil
}

func getIsuGraphErrorAction(ctx context.Context, a *agent.Agent, id string, query url.Values) (string, *http.Response, error) {
	path := fmt.Sprintf("/api/isu/%s/graph", id)
	rpath := getPathWithParams(path, query)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, rpath, []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getTrendAction(ctx context.Context, a *agent.Agent) (service.GetTrendResponse, *http.Response, error) {
	trend := service.GetTrendResponse{}
	reqUrl := "/api/trend"
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &trend, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}

	//TODO: バリデーション
	// res, text, err := reqJSONResError(ctx, a, http.MethodPost, reqUrl, bytes.NewReader(body), []int{http.StatusNotFound, http.StatusBadRequest})
	// if err != nil {
	// 	return "", nil, err
	// }

	return trend, res, nil
}

func browserGetHomeAction(ctx context.Context, a *agent.Agent,
	virtualNowUnix int64,
	allowNotModified bool,
	validateIsu func(*http.Response, []*service.Isu) []error,
) ([]*service.Isu, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	isuList, hres, err := getIsuAction(ctx, a)
	if err != nil {
		errors = append(errors, err)
	}
	if isuList != nil {
		var wg sync.WaitGroup
		var errMutex sync.Mutex
		wg.Add(len(isuList))
		for _, isu := range isuList {
			go func(isu *service.Isu) {
				defer wg.Done()
				icon, _, err := getIsuIconAction(ctx, a, isu.JIAIsuUUID, allowNotModified)
				if err != nil {
					isu.Icon = nil
					errMutex.Lock()
					errors = append(errors, err)
					errMutex.Unlock()
				} else {
					isu.Icon = icon
				}
			}(isu)
		}
		wg.Wait()
		errors = append(errors, validateIsu(hres, isuList)...)
	}

	return isuList, errors
}

func browserGetRegisterAction(ctx context.Context, a *agent.Agent) []error {
	// TODO: 静的ファイルのGET

	errors := []error{}
	return errors
}

func browserGetAuthAction(ctx context.Context, a *agent.Agent) []error {
	// TODO: 静的ファイルのGET

	errors := []error{}
	return errors
}

func browserGetIsuDetailAction(ctx context.Context, a *agent.Agent, id string,
	allowNotModified bool,
) (*service.Isu, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここはISU個別ページから遷移してきたならすでに持ってるからリクエストしない(変えてもいいけどフロントが不思議な実装になる)
	isu, _, err := getIsuIdAction(ctx, a, id)
	if err != nil {
		errors = append(errors, err)
	}
	if isu != nil {
		icon, _, err := getIsuIconAction(ctx, a, id, allowNotModified)
		if err != nil {
			isu.Icon = nil
			errors = append(errors, err)
		} else {
			isu.Icon = icon
		}

		return isu, errors
	}
	return nil, errors
}

func browserGetIsuConditionAction(ctx context.Context, a *agent.Agent, id string, req service.GetIsuConditionRequest,
	validateCondition func(*http.Response, []*service.GetIsuConditionResponse) []error,
) (*service.Isu, []*service.GetIsuConditionResponse, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここはISU個別ページから遷移してきたならすでに持ってるからリクエストしない(変えてもいいけどフロントが不思議な実装になる)
	isu, _, err := getIsuIdAction(ctx, a, id)
	if err != nil {
		errors = append(errors, err)
	}
	conditions, hres, err := getIsuConditionAction(ctx, a, id, req)
	if err != nil {
		errors = append(errors, err)
	} else {
		errors = append(errors, validateCondition(hres, conditions)...)
	}
	return isu, conditions, errors
}

func browserGetIsuGraphAction(ctx context.Context, a *agent.Agent, id string, date int64,
	validateGraph func(*http.Response, service.GraphResponse) []error,
) (*service.Isu, service.GraphResponse, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここはISU個別ページから遷移してきたならすでに持ってるからリクエストしない(変えてもいいけどフロントが不思議な実装になる)
	isu, _, err := getIsuIdAction(ctx, a, id)
	if err != nil {
		errors = append(errors, err)
	}
	req := service.GetGraphRequest{Date: date}
	graph, res, err := getIsuGraphAction(ctx, a, id, req)
	if err != nil {
		errors = append(errors, err)
	} else {
		errors = append(errors, validateGraph(res, graph)...)
	}
	return isu, graph, errors
}

func BrowserAccess(ctx context.Context, a *agent.Agent, rpath string, page PageType) error {
	req, err := a.GET(rpath)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	res, err := a.Do(ctx, req)
	if err != nil {
		return failure.NewError(ErrHTTP, err)
	}
	if err := verifyStatusCode(res, http.StatusOK); err != nil {
		if err := verifyStatusCode(res, http.StatusNotModified); err != nil {
			return failure.NewError(ErrInvalidStatusCode, err)
		}
	}

	resources, err := a.ProcessHTML(ctx, res, res.Body)
	if err != nil {
		return failure.NewError(ErrCritical, err)
	}
	// resourceの検証
	errs := verifyResources(page, res, resources)
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}
