package scenario

// verify.go
// 各種検証のユーティリティ関数
// ErrBadResponseのあたりの書式チェックと、
// シナリオのstructがあれば文脈無しで検証できるもの

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
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
func verify4xxError(res *http.Response, text string, expectedText string, expectedCode int) error {
	if res.StatusCode != expectedCode {
		return errorInvalidStatusCode(res, expectedCode)
	}
	if text != expectedText {
		return errorMissmatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", text, expectedCode)
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
				exp.JIACatalogID == isu.JIACatalogID &&
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

//TODO:
// func verifyCatalog(res *http.Response, expected *model.Catalog, catalog *service.Catalog) []error {
// 	errs := []error{}
// 	return errs
// }

//
//mustExistUntil: この値以下のtimestampを持つものは全て反映されているべき
func verifyIsuConditions(res *http.Response,
	targetUser *model.User, targetIsuUUID string, request *service.GetIsuConditionRequest,
	backendData []*service.GetIsuConditionResponse, mustExistUntil int64) error {

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
	var baseIter model.IsuConditionIterator
	if targetIsuUUID != "" {
		targetIsu := targetUser.IsuListByID[targetIsuUUID]
		iterTmp := targetIsu.Conditions.LowerBound(filter, request.CursorEndTime, request.CursorJIAIsuUUID)
		baseIter = &iterTmp
	} else {
		iterTmp := targetUser.Conditions.LowerBound(filter, request.CursorEndTime, request.CursorJIAIsuUUID)
		baseIter = &iterTmp
	}

	//backendDataの先頭からチェック
	var lastSort model.IsuConditionCursor
	for i, c := range backendData {
		nowSort := model.IsuConditionCursor{TimestampUnix: c.Timestamp, OwnerID: c.JIAIsuUUID}
		if i != 0 && !nowSort.Less(&lastSort) {
			return errorInvalid(res, "整列順が正しくありません")
		}

		var expected *model.IsuCondition
		for {
			expected = baseIter.Prev()
			if expected == nil {
				return errorMissmatch(res, "POSTに成功していない時刻のデータが返されました")
			}

			if expected.TimestampUnix == c.Timestamp && expected.OwnerID == c.JIAIsuUUID {
				break //ok
			}

			if mustExistUntil < expected.TimestampUnix {
				//反映されていないことが許可されているので、無視して良い
				continue
			}
			return errorMissmatch(res, "POSTに成功していない時刻のデータが返されました")
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
			c.JIAIsuUUID != expected.OwnerID ||
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
