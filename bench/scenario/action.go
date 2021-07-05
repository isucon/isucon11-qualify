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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

//Action

// ==============================initialize==============================

func initializeAction(ctx context.Context, a *agent.Agent) (*service.InitializeResponse, []error) {
	errors := []error{}

	//リクエスト
	req, err := a.POST("/initialize", nil)
	if err != nil {
		logger.AdminLogger.Panic(err)
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
		err = errorBadResponse(res, "利用言語(language)が設定されていません")
		errors = append(errors, err)
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
		userID2, err := model.MakeRandomUserID()
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
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
	return res, err
}

func signoutErrorAction(ctx context.Context, a *agent.Agent) (string, *http.Response, error) {
	res, text, err := reqNoContentResError(ctx, a, http.MethodPost, "/api/signout", []int{http.StatusUnauthorized})
	if err != nil {
		return "", nil, err
	}
	return text, res, err
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
		return "", nil, err
	}
	return resBody, res, nil
}

func getIsuAction(ctx context.Context, a *agent.Agent) ([]*service.Isu, *http.Response, error) {
	var isuList []*service.Isu
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, "/api/isu", nil, &isuList, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return isuList, res, nil
}

func getIsuErrorAction(ctx context.Context, a *agent.Agent) (string, *http.Response, error) {
	res, resBody, err := reqJSONResError(ctx, a, http.MethodGet, "/api/isu", nil, []int{http.StatusUnauthorized})
	if err != nil {
		return "", nil, err
	}
	return resBody, res, nil
}

func postIsuAction(ctx context.Context, a *agent.Agent, req service.PostIsuRequest) (*service.Isu, *http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	isu := &service.Isu{}
	res, err := reqJSONResJSON(ctx, a, http.MethodPost, "/api/isu", bytes.NewReader(body), isu, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return isu, res, nil
}

func postIsuErrorAction(ctx context.Context, a *agent.Agent, req interface{}) (string, *http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	res, text, err := reqJSONResError(ctx, a, http.MethodPost, "/api/isu", bytes.NewReader(body), []int{http.StatusBadRequest, http.StatusConflict, http.StatusUnauthorized, http.StatusNotFound})
	if err != nil {
		return "", nil, err
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

func putIsuAction(ctx context.Context, a *agent.Agent, id string, req service.PutIsuRequest) (*service.Isu, *http.Response, error) {
	isu := &service.Isu{}
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	reqUrl := fmt.Sprintf("/api/isu/%s", id)
	res, err := reqJSONResJSON(ctx, a, http.MethodPut, reqUrl, bytes.NewReader(body), &isu, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return isu, res, nil
}

func putIsuErrorAction(ctx context.Context, a *agent.Agent, id string, req service.PutIsuRequest) (string, *http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	reqUrl := fmt.Sprintf("/api/isu/%s", id)
	res, text, err := reqJSONResError(ctx, a, http.MethodPut, reqUrl, bytes.NewReader(body), []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func deleteIsuAction(ctx context.Context, a *agent.Agent, id string) (*http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s", id)
	res, err := reqNoContentResNoContent(ctx, a, http.MethodDelete, reqUrl, []int{http.StatusNoContent})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func deleteIsuErrorAction(ctx context.Context, a *agent.Agent, id string) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s", id)
	res, text, err := reqNoContentResError(ctx, a, http.MethodDelete, reqUrl, []int{http.StatusUnauthorized, http.StatusNotFound})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getIsuIconAction(ctx context.Context, a *agent.Agent, id string) ([]byte, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/icon", id)
	res, image, err := reqNoContentResPng(ctx, a, http.MethodGet, reqUrl, []int{http.StatusOK})
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

// TODO: 画像は取り扱いをどうしたらいいかわかんないからとりあえず io.Reader で
func putIsuIconAction(ctx context.Context, a *agent.Agent, id string, image io.Reader) (*http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/icon", id)
	res, err := reqPngResNoContent(ctx, a, http.MethodPut, reqUrl, image, []int{http.StatusOK})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func putIsuIconErrorAction(ctx context.Context, a *agent.Agent, id string, image io.Reader) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/icon", id)
	res, text, err := reqPngResError(ctx, a, http.MethodPut, reqUrl, image, []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getIsuSearchAction(ctx context.Context, a *agent.Agent, req service.GetIsuSearchRequest) ([]*service.Isu, *http.Response, error) {
	reqUrl := getIsuSearchRequestParams(req)
	var isuList []*service.Isu
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &isuList, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return isuList, res, nil
}

func getIsuSearchErrorAction(ctx context.Context, a *agent.Agent, req service.GetIsuSearchRequest) (string, *http.Response, error) {
	reqUrl := getIsuSearchRequestParams(req)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusUnauthorized, http.StatusNotFound})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getIsuSearchRequestParams(req service.GetIsuSearchRequest) string {
	targetURLStr := "/api/isu/search"
	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}

	q := url.Values{}
	if req.Name != nil {
		q.Set("name", *req.Name)
	}
	if req.CatalogName != nil {
		q.Set("catalog_name", *req.CatalogName)
	}
	if req.CatalogTags != nil {
		q.Set("catalog_tags", *req.CatalogTags)
	}
	if req.Character != nil {
		q.Set("character", *req.Character)
	}
	if req.MinLimitWeight != nil {
		q.Set("min_limit_weight", fmt.Sprint(*req.MinLimitWeight))
	}
	if req.MaxLimitWeight != nil {
		q.Set("max_limit_weight", fmt.Sprint(*req.MaxLimitWeight))
	}
	targetURL.RawQuery = q.Encode()
	return targetURL.String()
}

func getCatalogAction(ctx context.Context, a *agent.Agent, id string) (*service.Catalog, *http.Response, error) {
	catalog := &service.Catalog{}
	reqUrl := fmt.Sprintf("/api/catalog/%s", id)
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &catalog, []int{http.StatusOK})
	if err != nil {
		return catalog, nil, err
	}
	return catalog, res, nil
}

func getCatalogErrorAction(ctx context.Context, a *agent.Agent, id string) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/catalog/%s", id)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusUnauthorized, http.StatusNotFound})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func postIsuConditionAction(ctx context.Context, a *agent.Agent, id string, req service.PostIsuConditionRequest) (*http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/condition", id)
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	res, err := reqJSONResNoContent(ctx, a, http.MethodPost, reqUrl, bytes.NewReader(body), []int{http.StatusCreated})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func postIsuConditionErrorAction(ctx context.Context, a *agent.Agent, id string, req service.PostIsuConditionRequest) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/condition", id)
	body, err := json.Marshal(req)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	res, text, err := reqJSONResError(ctx, a, http.MethodPost, reqUrl, bytes.NewReader(body), []int{http.StatusNotFound, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getIsuConditionAction(ctx context.Context, a *agent.Agent, id string, req service.GetIsuConditionRequest) ([]*service.GetIsuConditionResponse, *http.Response, error) {
	reqUrl := getIsuConditionRequestParams(fmt.Sprintf("/api/condition/%s?", id), req)
	conditions := []*service.GetIsuConditionResponse{}
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &conditions, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return conditions, res, nil
}

func getIsuConditionErrorAction(ctx context.Context, a *agent.Agent, id string, req service.GetIsuConditionRequest) (string, *http.Response, error) {
	reqUrl := getIsuConditionRequestParams(fmt.Sprintf("/api/condition/%s?", id), req)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusNotFound, http.StatusUnauthorized})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func getConditionAction(ctx context.Context, a *agent.Agent, req service.GetIsuConditionRequest) ([]*service.GetIsuConditionResponse, *http.Response, error) {
	reqUrl := getIsuConditionRequestParams("/api/condition?", req)
	conditions := []*service.GetIsuConditionResponse{}
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &conditions, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return conditions, res, nil
}

func getConditionErrorAction(ctx context.Context, a *agent.Agent, req service.GetIsuConditionRequest) (string, *http.Response, error) {
	reqUrl := getIsuConditionRequestParams("/api/condition?", req)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusNotFound, http.StatusUnauthorized})
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
	q.Set("cursor_end_time", fmt.Sprint(req.CursorEndTime))
	if req.CursorJIAIsuUUID != "" {
		q.Set("cursor_jia_isu_uuid", req.CursorJIAIsuUUID)
	}
	q.Set("condition_level", req.ConditionLevel)
	if req.Limit != nil {
		q.Set("limit", fmt.Sprint(*req.Limit))
	}
	targetURL.RawQuery = q.Encode()
	return targetURL.String()
}

func getIsuGraphAction(ctx context.Context, a *agent.Agent, id string, date uint64) ([]*service.GraphResponse, *http.Response, error) {
	graph := []*service.GraphResponse{}
	reqUrl := fmt.Sprintf("/api/isu/%s/graph?date=%d", id, date)
	res, err := reqJSONResJSON(ctx, a, http.MethodGet, reqUrl, nil, &graph, []int{http.StatusOK})
	if err != nil {
		return nil, nil, err
	}
	return graph, res, nil
}

func getIsuGraphErrorAction(ctx context.Context, a *agent.Agent, id string, date uint64) (string, *http.Response, error) {
	reqUrl := fmt.Sprintf("/api/isu/%s/graph?date=%d", id, date)
	res, text, err := reqNoContentResError(ctx, a, http.MethodGet, reqUrl, []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusBadRequest})
	if err != nil {
		return "", nil, err
	}
	return text, res, nil
}

func browserGetHomeAction(ctx context.Context, a *agent.Agent,
	validateIsu func(*http.Response, []*service.Isu) []error,
	validateCondition func(*http.Response, []*service.GetIsuConditionResponse) []error,
) ([]*service.Isu, []*service.GetIsuConditionResponse, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここ以下は多分並列
	isuList, hres, err := getIsuAction(ctx, a)
	if err != nil {
		errors = append(errors, err)
	}
	if isuList != nil {
		// TODO: ここ以下は多分並列
		for _, isu := range isuList {
			icon, _, err := getIsuIconAction(ctx, a, isu.JIAIsuUUID)
			if err != nil {
				errors = append(errors, err)
			}
			isu.Icon = icon
		}
		errors = append(errors, validateIsu(hres, isuList)...)
	}

	conditions, hres, err := getConditionAction(ctx, a, service.GetIsuConditionRequest{CursorEndTime: uint64(time.Now().Unix()), CursorJIAIsuUUID: "z", ConditionLevel: "critical,warning,info"})
	if err != nil {
		errors = append(errors, err)
	} else {
		errors = append(errors, validateCondition(hres, conditions)...)
	}
	return isuList, conditions, errors
}

func browserGetSearchAction(ctx context.Context, a *agent.Agent, req service.GetIsuSearchRequest) ([]*service.Isu, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	isuList, _, err := getIsuSearchAction(ctx, a, req)
	if err != nil {
		errors = append(errors, err)
	}
	if isuList != nil {
		// TODO: ここ以下は多分並列
		for _, isu := range isuList {
			icon, _, err := getIsuIconAction(ctx, a, isu.JIAIsuUUID)
			if err != nil {
				errors = append(errors, err)
			}
			isu.Icon = icon
		}
	}
	return isuList, errors
}

func browserGetConditionsAction(ctx context.Context, a *agent.Agent, req service.GetIsuConditionRequest) ([]*service.GetIsuConditionResponse, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここ以下は多分並列
	conditions, _, err := getConditionAction(ctx, a, req)
	if err != nil {
		errors = append(errors, err)
	}
	return conditions, errors
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
	validateCatalog func(*http.Response, *service.Catalog) []error,
) (*service.Isu, *service.Catalog, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここはISU個別ページから遷移してきたならすでに持ってるからリクエストしない(変えてもいいけどフロントが不思議な実装になる)
	isu, _, err := getIsuIdAction(ctx, a, id)
	if err != nil {
		errors = append(errors, err)
	}
	if isu != nil {
		// TODO: ここ以下は多分並列
		icon, _, err := getIsuIconAction(ctx, a, id)
		if err != nil {
			errors = append(errors, err)
		}
		isu.Icon = icon

		catalog, hres, err := getCatalogAction(ctx, a, isu.JIACatalogID)
		if err != nil {
			errors = append(errors, err)
		} else {
			errors = append(errors, validateCatalog(hres, catalog)...)
		}
		return isu, catalog, errors
	}
	return nil, nil, errors
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

func browserGetIsuGraph(ctx context.Context, a *agent.Agent, id string, date uint64) (*service.Isu, []*service.GraphResponse, []error) {
	// TODO: 静的ファイルのGET

	errors := []error{}
	// TODO: ここはISU個別ページから遷移してきたならすでに持ってるからリクエストしない(変えてもいいけどフロントが不思議な実装になる)
	isu, _, err := getIsuIdAction(ctx, a, id)
	if err != nil {
		errors = append(errors, err)
	}
	graph, _, err := getIsuGraphAction(ctx, a, id, date)
	if err != nil {
		errors = append(errors, err)
	}
	return isu, graph, errors
}
