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
		if exp.JIACatalogID == isu.JIACatalogID {
			if exp.Character == isu.Character &&
				exp.JIAIsuUUID == isu.JIAIsuUUID &&
				exp.Name == isu.Name {
				//TODO: iconの検証

			} else {
				errs = append(errs, errorMissmatch(res, "%d番目の椅子の情報が異なります: ID=%s", i+1, isu.JIACatalogID))
			}
		} else {
			errs = append(errs, errorMissmatch(res, "%d番目の椅子が異なります: ID=%s (expected=%s)", i+1, isu.JIACatalogID, exp.JIACatalogID))
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
	base *model.IsuConditionArray, filter model.ConditionLevel, cursor model.IsuConditionCursor, isuMap map[string]*model.Isu,
	backendData []*service.GetIsuConditionResponse, mustExistUntil int64) error {

	//expectedの開始位置を探す()
	baseIter := base.End(filter)
	baseIter.UpperBoundIsuConditionIndex(cursor.TimestampUnix, cursor.OwnerID)

	//backendDataの先頭からチェック
	lastSort := model.IsuConditionCursor{TimestampUnix: backendData[0].Timestamp + 1, OwnerID: ""}
	for _, c := range backendData {
		nowSort := model.IsuConditionCursor{TimestampUnix: c.Timestamp, OwnerID: c.JIAIsuUUID}
		if !nowSort.Less(&lastSort) {
			return errorInvalid(res, "整列順が正しくありません")
		}

		var expected *model.IsuCondition
		for {
			expected = baseIter.Prev()
			if expected == nil {
				return errorMissmatch(res, "存在しないはずのデータが返されました")
			}

			if expected.TimestampUnix == c.Timestamp {
				break //ok
			}

			if expected.TimestampUnix <= mustExistUntil {
				return errorMissmatch(res, "データが足りません")
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
			c.JIAIsuUUID != expected.OwnerID ||
			c.Message != expected.Message ||
			c.IsuName != isuMap[expected.OwnerID].Name {
			return errorMissmatch(res, "データが正しくありません")
		}

		lastSort = nowSort
	}

	return nil
}
