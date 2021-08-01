package scenario

// verify.go
// 各種検証のユーティリティ関数
// ErrBadResponseのあたりの書式チェックと、
// シナリオのstructがあれば文脈無しで検証できるもの

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"

	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

//汎用関数

func verifyStatusCodes(res *http.Response, allowedStatusCodes []int) error {
	invalidStatusCode := true
	for _, c := range allowedStatusCodes {
		if res.StatusCode == c {
			invalidStatusCode = false
			break
		}
	}
	if invalidStatusCode {
		return errorInvalidStatusCodes(res, allowedStatusCodes)
	}
	return nil
}

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
func verify4xxError(res *http.Response, text string, expectedText string, expectedCode int) error {
	if res.StatusCode != expectedCode {
		return errorInvalidStatusCode(res, expectedCode)
	}
	if text != expectedText {
		return errorMissmatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", text, expectedText)
	}
	return nil
}

// 文脈無しで検証できるもの

func verifyNotSignedIn(res *http.Response, text string) error {
	expected := "you are not signed in"
	return verify4xxError(res, text, expected, http.StatusUnauthorized)
}

// TODO: 統一され次第消す
func verifyNotSignedInTODO(res *http.Response, text string) error {
	expected := "you are not sign in"
	return verify4xxError(res, text, expected, http.StatusUnauthorized)
}

func verifyBadReqBody(res *http.Response, text string) error {
	expected := "bad request body"
	return verify4xxError(res, text, expected, http.StatusBadRequest)
}

func verifyIsuNotFound(res *http.Response, text string) error {
	expected := "isu not found"
	return verify4xxError(res, text, expected, http.StatusNotFound)
}

//データ整合性チェック

func verifyIsuOrderByCreatedAt(res *http.Response, expectedReverse []*model.Isu, isuList []*service.Isu) []error {
	errs := []error{}
	length := len(expectedReverse)
	if length != len(isuList) {
		errs = append(errs, errorMissmatch(res, "椅子の数が異なります"))
		return errs
	}
	for i, isu := range isuList {
		exp := expectedReverse[length-1-i]
		if exp.JIAIsuUUID == isu.JIAIsuUUID {
			if exp.Character == isu.Character &&
				exp.Name == isu.Name {
				//TODO: iconの検証

			} else {
				errs = append(errs, errorMissmatch(res, "%d番目の椅子の情報が異なります: ID=%s", i+1, isu.JIAIsuUUID))
			}
		} else {
			errs = append(errs, errorMissmatch(res, "%d番目の椅子が異なります: ID=%s (expected=%s)", i+1, isu.JIAIsuUUID, exp.JIAIsuUUID))
		}
	}

	return errs
}

//mustExistUntil: この値以下のtimestampを持つものは全て反映されているべき
func verifyIsuConditions(res *http.Response,
	targetUser *model.User, targetIsuUUID string, request *service.GetIndividualIsuConditionRequest,
	backendData []*service.GetIsuConditionResponse) error {

	//limitを超えているかチェック
	var limit int
	if request.Limit != nil {
		limit = int(*request.Limit)
	} else {
		limit = conditionLimit
	}
	if limit < len(backendData) {
		return errorInvalid(res, "要素数が正しくありません")
	}
	//レスポンス側のstartTimeのチェック
	if request.StartTime != nil && len(backendData) != 0 && backendData[len(backendData)-1].Timestamp < *request.StartTime {
		return errorInvalid(res, "データが正しくありません")
	}

	//expectedの開始位置を探す
	filter := model.ConditionLevelNone
	for _, level := range strings.Split(request.ConditionLevel, ",") {
		switch level[0] {
		case 'i':
			filter |= model.ConditionLevelInfo
		case 'w':
			filter |= model.ConditionLevelWarning
		case 'c':
			filter |= model.ConditionLevelCritical
		}
	}

	targetIsu := targetUser.IsuListByID[targetIsuUUID]
	iterTmp := targetIsu.Conditions.LowerBound(filter, request.CursorEndTime, targetIsuUUID)
	baseIter := &iterTmp

	//backendDataは新しい順にソートされているはずなので、先頭からチェック
	var lastSort model.IsuConditionCursor
	for i, c := range backendData {
		//backendDataが新しい順にソートされていることの検証
		nowSort := model.IsuConditionCursor{TimestampUnix: c.Timestamp, OwnerIsuUUID: c.JIAIsuUUID}
		if i != 0 && !nowSort.Less(&lastSort) {
			return errorInvalid(res, "整列順が正しくありません")
		}

		var expected *model.IsuCondition
		for {
			expected = baseIter.Prev()
			if expected == nil {
				return errorMissmatch(res, "POSTに成功していない時刻のデータが返されました")
			}

			if expected.TimestampUnix == c.Timestamp && expected.OwnerIsuUUID == c.JIAIsuUUID {
				break //ok
			}

			if expected.TimestampUnix < c.Timestamp {
				return errorMissmatch(res, "POSTに成功していない時刻のデータが返されました")
			}
		}

		//等価チェック
		expectedCondition := fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v",
			expected.IsDirty,
			expected.IsOverweight,
			expected.IsBroken,
		)
		var expectedConditionLevelStr string
		warnCount := 0
		if expected.IsDirty {
			warnCount++
		}
		if expected.IsOverweight {
			warnCount++
		}
		if expected.IsBroken {
			warnCount++
		}
		switch warnCount {
		case 0:
			expectedConditionLevelStr = "info"
		case 1, 2:
			expectedConditionLevelStr = "warning"
		case 3:
			expectedConditionLevelStr = "critical"
		}
		if c.Condition != expectedCondition ||
			c.ConditionLevel != expectedConditionLevelStr ||
			c.IsSitting != expected.IsSitting ||
			c.JIAIsuUUID != expected.OwnerIsuUUID ||
			c.Message != expected.Message ||
			c.IsuName != targetUser.IsuListByID[c.JIAIsuUUID].Name {
			return errorMissmatch(res, "データが正しくありません")
		}
		lastSort = nowSort
	}

	//limitの検証
	if len(backendData) < limit && baseIter.Prev() != nil {
		prev := baseIter.Prev()
		if prev != nil && request.StartTime != nil && *request.StartTime <= prev.TimestampUnix {
			return errorInvalid(res, "要素数が正しくありません")
		}
	}

	return nil
}

// TODO: 実装する
func verifyAllConditions(res *http.Response,
	targetUser *model.User, request *service.GetIsuConditionRequest,
	backendData []*service.GetIsuConditionResponse, mustExistUntil int64) error {
	return nil
}
func joinURL(base *url.URL, target string) string {
	b := *base
	t, _ := url.Parse(target)
	u := b.ResolveReference(t).String()
	return u
}

// TODO: vendor.****.jsで取得処理が記述されているlogo_white, logo_orangeも取得できてない
func verifyResources(page string, res *http.Response, resources agent.Resources) []error {
	base := res.Request.URL.String()

	faviconSvg := resourcesMap["/favicon.svg"]
	indexCss := resourcesMap["/index.css"]
	indexJs := resourcesMap["/index.js"]
	//logoOrange := resourcesMap["/logo_orange.svg"]
	//logoWhite := resourcesMap["/logo_white.svg"]
	vendorJs := resourcesMap["/vendor.js"]

	var checks []error
	switch page {
	case "/signup":
		checks = []error{
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+faviconSvg)], faviconSvg),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexCss)], indexCss),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexJs)], indexJs),
			//errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+logoWhite)], logoWhite),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+vendorJs)], vendorJs),
		}
	case "/condition":
		checks = []error{
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+faviconSvg)], faviconSvg),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexCss)], indexCss),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexJs)], indexJs),
			//errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+logoWhite)], logoWhite),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+vendorJs)], vendorJs),
		}
	case "/isu":
		checks = []error{
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+faviconSvg)], faviconSvg),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexCss)], indexCss),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexJs)], indexJs),
			//errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+logoWhite)], logoWhite),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+vendorJs)], vendorJs),
		}
	case "/register":
		checks = []error{
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+faviconSvg)], faviconSvg),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexCss)], indexCss),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexJs)], indexJs),
			//errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+logoWhite)], logoWhite),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+vendorJs)], vendorJs),
		}
	case "/login":
		checks = []error{
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+faviconSvg)], faviconSvg),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexCss)], indexCss),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+indexJs)], indexJs),
			//errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+logoOrange)], logoOrange),
			//errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+logoWhite)], logoWhite),
			errorChecksum(base, resources[joinURL(res.Request.URL, "/assets"+vendorJs)], vendorJs),
		}
	}
	errs := []error{}
	for _, err := range checks {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func errorChecksum(base string, resource *agent.Resource, name string) error {
	if resource == nil {
		logger.AdminLogger.Printf("resource not found: %s on %s\n", name, base)
		return errorCheckSum("期待するリソースが読み込まれませんでした: %s", name)
	}

	if resource.Error != nil {
		var nerr net.Error
		if failure.As(resource.Error, &nerr) {
			if nerr.Timeout() || nerr.Temporary() {
				return nerr
			}
		}
		return errorCheckSum("リソースの取得に失敗しました: %s: %v", name, resource.Error)
	}

	res := resource.Response
	defer res.Body.Close()
	if res.StatusCode == 304 {
		return nil
	}

	if err := verifyStatusCode(res, http.StatusOK); err != nil {
		return err
	}

	// md5でリソースの比較
	path := res.Request.URL.Path
	expected := resourcesHash[path]
	if expected == "" {
		return nil
	}
	hash := md5.New()
	if _, err := io.Copy(hash, res.Body); err != nil {
		logger.AdminLogger.Printf("resource checksum: %v", err)
		return errorCheckSum("リソースの取得に失敗しました: %s", path)
	}
	actual := fmt.Sprintf("%x", hash.Sum(nil))
	if expected != actual {
		return errorCheckSum("期待するチェックサムと一致しません: %s", path)
	}
	return nil
}

func verifyGraph(
	res *http.Response, targetUser *model.User, targetIsuUUID string,
	getGraphReq *service.GetGraphRequest,
	getGraphResp service.GraphResponse) error {

	// graphResp の配列は必ず 24 つ (24時間分) である
	if len(getGraphResp) != 24 {
		return errorInvalid(res, "要素数が正しくありません")
	}

	var lastStartAt int64
	// getGraphResp を逆順 (timestamp が新しい順) にloop
	for idxGraphResp := len(getGraphResp) - 1; idxGraphResp >= 0; idxGraphResp-- {
		graphOne := getGraphResp[idxGraphResp]

		// getGraphResp の要素が古い順に連続して並んでいることの検証
		if idxGraphResp != len(getGraphResp)-1 && !(graphOne.EndAt == lastStartAt) {
			return errorInvalid(res, "整列順が正しくありません")
		}
		lastStartAt = graphOne.StartAt

		// 特定の ISU における expected な conditions を新しい順に取得するイテレータを生成
		targetIsu := targetUser.IsuListByID[targetIsuUUID]
		filter := model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical
		baseIter := targetIsu.Conditions.LowerBound(filter, graphOne.EndAt, targetIsu.JIAIsuUUID)

		var conditionsBaseOfScore []*model.IsuCondition
		var lastSort model.IsuConditionCursor
		// graphOne.ConditionTimestamps を逆順 (timestamp が新しい順) に loop
		for idxTimestamps := len(graphOne.ConditionTimestamps) - 1; idxTimestamps >= 0; idxTimestamps-- {
			timestamp := graphOne.ConditionTimestamps[idxTimestamps]

			// graphOne.start_at <= graphOne.condition_timestamps < graphOne.end_at であることの検証
			if !(graphOne.StartAt <= timestamp && timestamp < graphOne.EndAt) {
				return errorInvalid(res, "condition_timestampsがstart_atからend_atの中に収まっていません")
			}

			// graphOne.ConditionTimestamps の要素が古い順に並んでいることの検証
			nowSort := model.IsuConditionCursor{TimestampUnix: timestamp}
			if idxTimestamps != len(graphOne.ConditionTimestamps)-1 && !nowSort.Less(&lastSort) {
				return errorInvalid(res, "整列順が正しくありません")
			}
			lastSort = nowSort

			// graphOne.ConditionTimestamps[*] が expected に存在することの検証
			var expected *model.IsuCondition
			for {
				expected = baseIter.Prev()
				// 降順イテレータから得た expected が timestamp を追い抜いた ⇒ actual が expected に無いデータを返している
				if expected == nil || expected.TimestampUnix < timestamp {
					return errorMissmatch(res, "POSTに成功していない時刻のデータが返されました")
				}
				if expected.TimestampUnix == timestamp {
					// graphOne.ConditionTimestamps[n] から condition を取得
					conditionsBaseOfScore = append(conditionsBaseOfScore, expected)
					break //ok
				}
			}
		}

		// actual の data が空の場合 verify skip
		if graphOne.Data == nil {
			continue
		}

		// conditionsBaseOfScore から組み立てた data が actual と等値であることの検証
		expectedGraph := model.NewGraph(conditionsBaseOfScore)

		if graphOne.Data.Score != expectedGraph.Score() ||
			graphOne.Data.Sitting != expectedGraph.Sitting() ||
			graphOne.Data.Detail["is_broken"] != expectedGraph.IsBroken() ||
			graphOne.Data.Detail["is_dirty"] != expectedGraph.IsDirty() ||
			graphOne.Data.Detail["is_overweight"] != expectedGraph.IsOverweight() ||
			graphOne.Data.Detail["missing_data"] != expectedGraph.MissingData() {
			return errorMissmatch(res, "graphのデータが正しくありません")
		}
	}

	return nil
}
