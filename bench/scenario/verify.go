package scenario

// verify.go
// 各種検証のユーティリティ関数
// ErrBadResponseのあたりの書式チェックと、
// シナリオのstructがあれば文脈無しで検証できるもの

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	if res.StatusCode == http.StatusNotModified {
		return nil
	}
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
		return errorMismatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", text, expected)
	}
	return nil
}
func verify4xxError(res *http.Response, text string, expectedText string, expectedCode int) error {
	if res.StatusCode != expectedCode {
		return errorInvalidStatusCode(res, expectedCode)
	}
	if text != expectedText {
		return errorMismatch(res, "エラーメッセージが不正確です: `%s` (expected: `%s`)", text, expectedText)
	}
	return nil
}

// 文脈無しで検証できるもの

func verifyNotSignedIn(res *http.Response, text string) error {
	expected := "you are not signed in"
	return verify4xxError(res, text, expected, http.StatusUnauthorized)
}

//データ整合性チェック

//Icon,LatestIsuConditionを除いたISUの整合性チェック
func verifyIsu(res *http.Response, expected *model.Isu, actual *service.Isu) error {
	if actual.JIAIsuUUID != expected.JIAIsuUUID {
		return errorMismatch(res, "椅子が異なります(expected %s, actual %s)", expected.JIAIsuUUID, actual.JIAIsuUUID)
	}
	if actual.ID != expected.ID ||
		actual.Character != expected.Character ||
		actual.Name != expected.Name {
		logger.AdminLogger.Printf("expected: { ID: %v, Character: %v, Name: %v }, actual: { ID: %v, Character: %v, Name: %v }", expected.ID, expected.Character, expected.Name, actual.ID, actual.Character, actual.Name)
		return errorMismatch(res, "椅子(JIA_ISU_UUID=%s)の情報が異なります", expected.JIAIsuUUID)
	}
	return nil
}

func verifyIsuIcon(expected *model.Isu, actual []byte, actualStatusCode int) error {
	if expected.ImageHash != md5.Sum(actual) {
		return failure.NewError(ErrMismatch, errorFormatWithURI(
			actualStatusCode, http.MethodGet, "/api/isu/"+expected.JIAIsuUUID+"/icon",
			"椅子のiconが異なります", //UUIDはpathでわかるので省略
		))
	}
	return nil
}

func verifyIsuList(res *http.Response, expectedReverse []*model.Isu, isuList []*service.Isu) ([]string, []error) {
	errs := []error{}
	var newConditionUUIDs []string
	length := len(expectedReverse)
	if length != len(isuList) {
		logger.AdminLogger.Printf("len(isuList). expected: %v, actual: %v", length, len(isuList))
		errs = append(errs, errorMismatch(res, "椅子の数が異なります"))
		return nil, errs
	}
	for i, isu := range isuList {

		expected := expectedReverse[length-1-i]

		//iconはエンドポイントが別なので、別枠で検証
		if isu.Icon == nil {
			//icon取得失敗(エラー追加済み)
		} else {
			err := verifyIsuIcon(expected, isu.Icon, isu.IconStatusCode)
			if err != nil {
				errs = append(errs, err)
			}
		}

		// isu の検証
		err := verifyIsu(res, expected, isu)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		// isu の検証 (latest_isu_condition)

		// isu.latest_isu_condition が nil &&  前回の latestIsuCondition の timestamp が初期値ならば
		if isu.LatestIsuCondition == nil && expected.LastReadConditionTimestamps[0] == 0 {
			// この ISU はまだ poster から condition を受け取っていないため skip
			continue
		}
		if isu.LatestIsuCondition == nil {
			continue
		}
		func() {
			expected.CondMutex.RLock()
			defer expected.CondMutex.RUnlock()

			baseIter := expected.Conditions.End(model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical)
			for {
				expectedCondition := baseIter.Prev()
				if expectedCondition == nil || expectedCondition.TimestampUnix < isu.LatestIsuCondition.Timestamp {
					errs = append(errs, errorMismatch(res, "%d番目の椅子 (JIA_ISU_UUID=%s) の情報が異なります: POSTに成功していない時刻のデータが返されました", i+1, isu.JIAIsuUUID))
					break
				}

				if expectedCondition.TimestampUnix == isu.LatestIsuCondition.Timestamp {
					// timestamp 以外の要素を検証
					if !(expected.JIAIsuUUID == isu.LatestIsuCondition.JIAIsuUUID &&
						expected.Name == isu.LatestIsuCondition.IsuName &&
						expectedCondition.IsSitting == isu.LatestIsuCondition.IsSitting &&
						expectedCondition.ConditionString() == isu.LatestIsuCondition.Condition &&
						expectedCondition.ConditionLevel.Equal(isu.LatestIsuCondition.ConditionLevel) &&
						expectedCondition.Message == isu.LatestIsuCondition.Message) {
						logger.AdminLogger.Printf(`expected: { JIAIsuUUID: %v, Name: %v, IsSitting: %v,`+
							`ConditionString: %v, ConditionLevel: %v, Message: %v, Timestamp: %v }`+
							`actual: { JAIsuUUID: %v, Name: %v, IsSitting: %v,`+
							`ConditionString: %v, ConditionLevel: %v, Message: %v, Timestamp: %v }`,
							expected.JIAIsuUUID, expected.Name, expectedCondition.IsSitting,
							expectedCondition.ConditionString(), expectedCondition.ConditionLevel, expectedCondition.Message, expectedCondition.TimestampUnix,
							isu.LatestIsuCondition.JIAIsuUUID, isu.LatestIsuCondition.IsuName, isu.LatestIsuCondition.IsSitting,
							isu.LatestIsuCondition.Condition, isu.LatestIsuCondition.ConditionLevel, isu.LatestIsuCondition.Message, isu.LatestIsuCondition.Timestamp)
						errs = append(errs, errorMismatch(res, "%d番目の椅子 (JIA_ISU_UUID=%s) の情報が異なります: latest_isu_conditionの内容が不正です", i+1, isu.JIAIsuUUID))
					} else if expected.LastReadConditionTimestamps[0] < isu.LatestIsuCondition.Timestamp {
						// もし前回の latestIsuCondition の timestamp より新しいならばカウンタをインクリメント
						// 更新はここではなく、conditionを見て加点したタイミングで更新
						newConditionUUIDs = append(newConditionUUIDs, isu.JIAIsuUUID)
					}
					break
				}
			}
		}()
	}
	return newConditionUUIDs, errs
}

//mustExistUntil: この値以下のtimestampを持つものは全て反映されているべき。副作用: IsuCondition の ReadTime を更新する。
func verifyIsuConditions(res *http.Response,
	targetUser *model.User, targetIsuUUID string, request *service.GetIsuConditionRequest,
	backendData service.GetIsuConditionResponseArray,
	mustExistTimestamps [service.ConditionLimit]int64,
	requestTimeUnix int64) error {

	//limitを超えているかチェック
	if service.ConditionLimit < len(backendData) {
		logger.AdminLogger.Printf("actual length: %v", len(backendData))
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
		default:
			logger.AdminLogger.Panicf("verifyIsuConditions: リクエストのクエリパラメータ condition_level が不正な値です: %v", level)
		}
	}

	targetIsu := targetUser.IsuListByID[targetIsuUUID]

	if err := func() error {
		// isu.Condition の read lock を取る
		targetIsu.CondMutex.RLock()
		defer targetIsu.CondMutex.RUnlock()

		conditions := targetIsu.Conditions
		iterTmp := conditions.LowerBound(filter, request.EndTime)
		baseIter := &iterTmp

		//backendDataは新しい順にソートされているはずなので、先頭からチェック
		var lastSort model.IsuConditionCursor
		for i, c := range backendData {
			//backendDataが新しい順にソートされていることの検証
			nowSort := model.IsuConditionCursor{TimestampUnix: c.Timestamp}
			if i != 0 && !nowSort.Less(&lastSort) {
				return errorInvalid(res, "整列順が正しくありません")
			}

			var expected *model.IsuCondition
			for {
				expected = baseIter.Prev()
				if expected == nil {
					return errorMismatch(res, "POSTに成功していない時刻のデータが返されました")
				}

				if expected.TimestampUnix == c.Timestamp {
					break //ok
				}

				if expected.TimestampUnix < c.Timestamp {
					logger.AdminLogger.Printf("actual timestamp: %v", c.Timestamp)
					return errorMismatch(res, "POSTに成功していない時刻のデータが返されました")
				}

				if request.StartTime != nil && c.Timestamp < *request.StartTime {
					logger.AdminLogger.Printf("actual timestamp: %v", c.Timestamp)
					return errorMismatch(res, "start_time に指定された時間より前のデータが返されました")
				}

				// GET /api/isu/:id/graph で見た ConditionDelayTime秒以上後に、その condition がないとき
				if expected.ReadTime+ConditionDelayTime < requestTimeUnix {
					return errorMismatch(res, "GET /api/condition/:jia_isu_uuid か GET /api/isu/:jia_isu_uuid/graph で確認された condition がありません")
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
				c.JIAIsuUUID != targetIsuUUID ||
				c.Message != expected.Message ||
				c.IsuName != targetIsu.Name {
				logger.AdminLogger.Printf("expected: {Condition: %v, ConditionLevel: %v, IsSitting: %v,	JIAIsuUUID: %v, Message: %v, IsuName: %v}, actual: {Condition: %v, ConditionLevel: %v, IsSitting: %v,	JIAIsuUUID: %v, Message: %v, IsuName: %v",
					expectedCondition, expectedConditionLevelStr, expected.IsSitting, targetIsuUUID, expected.Message, targetIsu.Name,
					c.Condition, c.ConditionLevel, c.IsSitting, c.JIAIsuUUID, c.Message, c.IsuName)
				return errorMismatch(res, "データが正しくありません")
			}
			lastSort = nowSort

			// GET /api/isu/:id/graph と連動してる読んだ時間を更新
			if expected.ReadTime > requestTimeUnix {
				expected.ReadTime = time.Now().Unix()
			}
		}

		// backendData が空のときもチェックする
		if len(backendData) == 0 {
			var expected *model.IsuCondition
			for {
				expected = baseIter.Prev()
				if expected == nil {
					break // ok
				}

				if request.StartTime != nil && *request.StartTime > expected.TimestampUnix {
					break // ok
				}

				// GET /api/isu/:id/graph で見た ConditionDelayTime秒以上後に、その condition がないとき
				if expected.ReadTime+ConditionDelayTime < requestTimeUnix {
					return errorMismatch(res, "GET /api/condition/:jia_isu_uuid か GET /api/isu/:jia_isu_uuid/graph で確認された condition がありません")
				}
			}
		}
		return nil
	}(); err != nil {
		return err
	}

	//mustExistTimestamps
	mustExistIndex := 0
	for request.EndTime <= mustExistTimestamps[mustExistIndex] {
		mustExistIndex++
		if service.ConditionLimit <= mustExistIndex {
			break
		}
	}
	for _, c := range backendData {
		if service.ConditionLimit <= mustExistIndex {
			break
		}

		if mustExistTimestamps[mustExistIndex] < c.Timestamp {
			continue
		}
		if mustExistTimestamps[mustExistIndex] == c.Timestamp {
			mustExistIndex++
			continue
		}
		return errorInvalid(res, "以前に存在を確認したデータが欠落しています")
	}
	if len(backendData) < service.ConditionLimit && mustExistIndex < service.ConditionLimit && mustExistTimestamps[mustExistIndex] != 0 {
		if request.StartTime == nil || *request.StartTime <= mustExistTimestamps[mustExistIndex] {
			//まだ表示されるべきデータが残っている
			logger.AdminLogger.Printf("actual length: %v", len(backendData))
			return errorInvalid(res, "limitに満たない件数のデータが返されました: 以前に存在を確認したデータが欠落しています")
		}
	}

	return nil
}

func verifyPrepareIsuConditions(res *http.Response,
	targetUser *model.User, targetIsuUUID string, request *service.GetIsuConditionRequest,
	backendData service.GetIsuConditionResponseArray) error {

	//limitを超えているかチェック
	if service.ConditionLimit < len(backendData) {
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
		default:
			logger.AdminLogger.Panicf("verifyPrepareIsuConditions: リクエストのクエリパラメータ condition_level が不正な値です: %v", level)
		}
	}

	targetIsu := targetUser.IsuListByID[targetIsuUUID]

	if err := func() error {
		// isu.Condition の read lock を取る
		targetIsu.CondMutex.RLock()
		defer targetIsu.CondMutex.RUnlock()

		iterTmp := targetIsu.Conditions.LowerBound(filter, request.EndTime)
		baseIter := &iterTmp

		//backendDataは新しい順にソートされているはずなので、先頭からチェック
		var lastSort model.IsuConditionCursor
		for i, c := range backendData {

			expected := baseIter.Prev()
			if expected == nil {
				return errorMismatch(res, "存在しないはずのデータが返却されています")
			}

			//backendDataが新しい順にソートされていることの検証
			nowSort := model.IsuConditionCursor{TimestampUnix: c.Timestamp}
			if i != 0 && !nowSort.Less(&lastSort) {
				return errorInvalid(res, "整列順が正しくありません")
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
				c.JIAIsuUUID != targetIsuUUID ||
				c.Message != expected.Message ||
				c.IsuName != targetIsu.Name ||
				c.Timestamp != expected.TimestampUnix {
				return errorMismatch(res, "データが正しくありません")
			}
			lastSort = nowSort
		}

		// limitの検証
		// response件数がlimitの数より少ないときは、bench側で条件に合うデータを更にもっていなければ正しい
		prev := baseIter.Prev()
		if len(backendData) < service.ConditionLimit && prev != nil {
			if request.StartTime != nil && *request.StartTime <= prev.TimestampUnix {
				return errorInvalid(res, "要素数が正しくありません")
			}
		}
		return nil
	}(); err != nil {
		return err
	}
	return nil
}

func joinURL(base *url.URL, target string) string {
	b := *base
	t, _ := url.Parse(target)
	u := b.ResolveReference(t).String()
	return u
}

func errorAssetChecksum(req *http.Request, res *http.Response, user AgentWithStaticCache, path string) error {
	if err := verifyStatusCodes(res, []int{http.StatusOK, http.StatusNotModified}); err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusNotModified {
		// cache の更新
		hasher := crc32.New(crc32.IEEETable)
		_, err := io.Copy(hasher, res.Body)
		if err != nil {
			return err
		}
		user.SetStaticCache(path, hasher.Sum32())
	}

	// crc32でリソースの比較
	expected := resourcesHash[path]
	if expected == "" {
		logger.AdminLogger.Panicf("意図していないpath(%s)のAssetResourceCheckを行っています。", path)
	}
	actualHash, exist := user.GetStaticCache(path, req)
	if !exist {
		if res.StatusCode != http.StatusNotModified {
			logger.AdminLogger.Panic("static cacheがありません")
		}
		return errorCheckSum("304 StatusNotModified を返却していますが cache がありません: %s", path)
	}
	actual := fmt.Sprintf("%x", actualHash)
	if expected != actual {
		return errorCheckSum("期待するチェックサムと一致しません: %s", path)
	}
	return nil
}

// 副作用: IsuCondition の ReadTime を更新する。
func verifyGraph(
	res *http.Response, targetUser *model.User, targetIsuUUID string,
	getGraphReq *service.GetGraphRequest,
	getGraphResp service.GraphResponse,
	requestTimeUnix int64) error {

	// graphResp の配列は必ず 24 つ (24時間分) である
	if len(getGraphResp) != 24 {
		logger.AdminLogger.Printf("actual length: %v", len(getGraphResp))
		return errorInvalid(res, "要素数が正しくありません")
	}

	reqDate := time.Unix(getGraphReq.Date, 0)
	startDate := truncateAfterHours(reqDate).Unix()
	endDate := truncateAfterHours(reqDate.Add(24 * time.Hour)).Unix()

	var lastStartAt int64
	// getGraphResp を逆順 (timestamp が新しい順) にloop
	for idxGraphResp := len(getGraphResp) - 1; idxGraphResp >= 0; idxGraphResp-- {
		graphOne := getGraphResp[idxGraphResp]

		// getGraphResp の要素が古い順に連続して並んでいることの検証
		if idxGraphResp != len(getGraphResp)-1 && !(graphOne.EndAt == lastStartAt) {
			return errorInvalid(res, "整列順が正しくありません")
		}
		lastStartAt = graphOne.StartAt

		// graphのデータが指定日内のものか検証
		if graphOne.StartAt < startDate || endDate < graphOne.EndAt {
			return errorInvalid(res, "グラフの日付が間違っています")
		}

		targetIsu := targetUser.IsuListByID[targetIsuUUID]
		var conditionsBaseOfScore []*model.IsuCondition

		if err := func() error {
			// isu.Condition の read lock を取る
			targetIsu.CondMutex.RLock()
			defer targetIsu.CondMutex.RUnlock()

			// 特定の ISU における expected な conditions を新しい順に取得するイテレータを生成
			conditions := targetIsu.Conditions
			filter := model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical
			baseIter := conditions.LowerBound(filter, graphOne.EndAt)

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
						logger.AdminLogger.Printf("actual timestamp: %v", timestamp)
						return errorMismatch(res, "POSTに成功していない時刻のデータが返されました")
					}
					if expected.TimestampUnix == timestamp {
						// graphOne.ConditionTimestamps[n] から condition を取得
						conditionsBaseOfScore = append(conditionsBaseOfScore, expected)

						// GET /api/condition/:id と連動してる読んだ時間を更新
						if expected.ReadTime > requestTimeUnix {
							expected.ReadTime = time.Now().Unix()
						}
						break //ok
					}

					// GET /api/condition/:id で見た ConditionDelayTime秒以上後に、その condition がないとき
					if expected.ReadTime+ConditionDelayTime < requestTimeUnix {
						logger.AdminLogger.Printf("must exist timestamp: %v", expected.TimestampUnix)
						return errorMismatch(res, "GET /api/condition/:jia_isu_uuid か GET /api/isu/:jia_isu_uuid/graph で確認された condition がありません")
					}
				}
			}

			// graphOne.ConditionTimestamps が空のときもチェックする
			if len(graphOne.ConditionTimestamps) == 0 {
				var expected *model.IsuCondition
				for {
					expected = baseIter.Prev()
					if expected == nil {
						break // ok
					}

					// 根拠 timestamps は StartAt <= timestamp < EndAt であるため
					if expected.TimestampUnix < graphOne.StartAt {
						break // ok
					}

					// GET /api/condition/:id で見た ConditionDelayTime秒以上後に、その condition がないとき
					if expected.ReadTime+ConditionDelayTime < requestTimeUnix {
						logger.AdminLogger.Printf("must exist timestamp: %v", expected.TimestampUnix)
						return errorMismatch(res, "GET /api/condition/:jia_isu_uuid か GET /api/isu/:jia_isu_uuid/graph で確認された condition がありません")
					}
				}
			}
			return nil
		}(); err != nil {
			return err
		}

		// conditionsBaseOfScore と graphOne.Data のどちらか一方が空のときはエラー
		if (len(conditionsBaseOfScore) == 0 || graphOne.Data == nil) && !(len(conditionsBaseOfScore) == 0 && graphOne.Data == nil) {
			return errorMismatch(res, "グラフの data と condition_timestamps のどちらか一方のみが null、あるいは空です。")
		}

		// actual の data が空の場合 verify skip
		if len(conditionsBaseOfScore) == 0 && graphOne.Data == nil {
			continue
		}

		// conditionsBaseOfScore から組み立てた data が actual と等値であることの検証
		expectedGraph := model.NewGraph(conditionsBaseOfScore)

		if !expectedGraph.Match(
			graphOne.Data.Score,
			graphOne.Data.Percentage.Sitting,
			graphOne.Data.Percentage.IsBroken,
			graphOne.Data.Percentage.IsDirty,
			graphOne.Data.Percentage.IsOverweight,
		) {
			return errorMismatch(res, "グラフのデータが正しくありません")
		}

	}
	return nil
}

// verifyTrend 内で利用する定数
var conditionList = []string{"info", "warning", "critical"}

func (s *Scenario) verifyTrend(
	ctx context.Context, res *http.Response,
	viewer *model.Viewer,
	trendResp service.GetTrendResponse,
	requestTime time.Time,
) (int, error) {

	// レスポンスの要素にある ISU の性格を格納するための set
	var characterSet model.IsuCharacterSet
	// レスポンスの要素にある ISU の ID を格納するための set
	isuIDSet := make(map[int]struct{}, 700)
	// 新規 conditions の数を格納するための変数
	var newConditionNum int

	// 前回 getTrend 実行時の ISU の数を取得
	previousConditionNum := viewer.NumOfIsu()

	for _, trendOne := range trendResp {

		character, err := model.NewIsuCharacter(trendOne.Character)
		if err != nil {
			return 0, errorInvalid(res, err.Error())
		}
		characterSet = characterSet.Append(character)

		for conditionEnum, conditions := range [][]service.TrendCondition{trendOne.Info, trendOne.Warning, trendOne.Critical} {
			conditionLevel := conditionList[conditionEnum]

			var lastConditionTimestamp int64
			for idx, condition := range conditions {

				// conditions が新しい順にソートされていることの検証
				if idx != 0 && !(condition.Timestamp <= lastConditionTimestamp) {
					return 0, errorInvalid(res, "整列順が正しくありません")
				}
				lastConditionTimestamp = condition.Timestamp

				// condition.ID から isu を取得する
				isu, ok := s.GetIsuFromID(condition.IsuID)
				if !ok {
					// 次のループでまた bench の知らない IsuID の ISU を見つけたら落とせるように
					if _, exist := isuIDSet[condition.IsuID]; exist {
						return 0, errorMismatch(res, "同じ ISU のコンディションが複数登録されています")
					}
					isuIDSet[condition.IsuID] = struct{}{}

					// POST /api/isu などのレスポンス待ちなためここで落とすことはできない
					continue
				}

				if err := func() error {
					// isu.Condition の read lock を取る
					isu.CondMutex.RLock()
					defer isu.CondMutex.RUnlock()

					// condition を最新順に取得するイテレータを生成
					filter := model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical
					conditions := isu.Conditions
					baseIter := conditions.UpperBound(filter, condition.Timestamp)

					// condition.timestamp と condition.condition の値を検証
					expected := baseIter.Prev()
					if expected == nil || expected.TimestampUnix != condition.Timestamp {
						return errorMismatch(res, "POSTに成功していない時刻のデータが返されました")
					}
					if !expected.ConditionLevel.Equal(conditionLevel) {
						return errorMismatch(res, "コンディションレベルが正しくありません")
					}
					// 同じ isu の condition が複数返されてないことの検証
					if _, exist := isuIDSet[condition.IsuID]; exist {
						return errorMismatch(res, "同じ ISU のコンディションが複数登録されています")
					}
					isuIDSet[condition.IsuID] = struct{}{}

					// 該当 condition が新規のものである場合はキャッシュを更新
					if !viewer.ConditionAlreadyVerified(condition.IsuID, condition.Timestamp) {
						// 該当 condition が以前のものよりも昔の timestamp で無いことの検証
						if !viewer.ConditionIsUpdated(condition.IsuID, condition.Timestamp) {
							return errorMismatch(res, "以前の取得結果よりも古いタイムスタンプのコンディションが返されています")
						}
						viewer.SetVerifiedCondition(condition.IsuID, condition.Timestamp)
						// 一秒前(仮想時間で16時間40分以上前)よりあとのものならカウンタをインクリメント
						if condition.Timestamp > s.ToVirtualTime(requestTime.Add(-2*time.Second)).Unix() {
							newConditionNum += 1
						}
					}

					return nil
				}(); err != nil {
					return 0, err
				}
			}
		}
	}
	// characterSet の検証
	if !characterSet.IsFull() {
		return 0, errorInvalid(res, "全ての性格のトレンドが取得できていません")
	}
	// trend のレスポンスに入っている ISU の数が expected な数以上あることの検証
	if !(len(isuIDSet) >= previousConditionNum) {
		return 0, errorInvalid(res, "ISU の個数が不足しています")
	}
	return newConditionNum, nil
}
func (s *Scenario) verifyPrepareTrend(
	res *http.Response,
	viewer *model.Viewer,
	trendResp service.GetTrendResponse,
) error {
	// レスポンスの要素にある ISU の性格を格納するための set
	var characterSet model.IsuCharacterSet
	// レスポンスの要素にある ISU の ID を格納するための set
	isuIDSet := make(map[int]struct{}, 700)

	for _, trendOne := range trendResp {

		character, err := model.NewIsuCharacter(trendOne.Character)
		if err != nil {
			return errorInvalid(res, err.Error())
		}
		characterSet = characterSet.Append(character)

		for conditionEnum, conditions := range [][]service.TrendCondition{trendOne.Info, trendOne.Warning, trendOne.Critical} {
			conditionLevel := conditionList[conditionEnum]

			var lastConditionTimestamp int64
			for idx, condition := range conditions {

				// conditions が新しい順にソートされていることの検証
				if idx != 0 && !(condition.Timestamp <= lastConditionTimestamp) {
					return errorInvalid(res, "整列順が正しくありません")
				}
				lastConditionTimestamp = condition.Timestamp

				// condition.ID から isu を取得する
				isu, ok := s.GetIsuFromID(condition.IsuID)
				// prepareでは初期データ以外のISUは存在しないはずなのでエラー
				if !ok {
					return errorInvalid(res, "存在しないはず ISU がレスポンスに含まれています")
				}

				if err := func() error {
					// isu.Condition の read lock を取る
					isu.CondMutex.RLock()
					defer isu.CondMutex.RUnlock()

					// condition を最新順に取得するイテレータを生成
					filter := model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical
					conditionArray := isu.Conditions
					baseIter := conditionArray.End(filter)

					// condition.timestamp と condition.condition の値を検証
					expected := baseIter.Prev()
					if expected == nil || expected.TimestampUnix != condition.Timestamp {
						return errorMismatch(res, "POSTに成功していない時刻のデータが返されました")
					}
					if !expected.ConditionLevel.Equal(conditionLevel) {
						return errorMismatch(res, "コンディションレベルが正しくありません")
					}
					// 同じ isu の condition が複数返されてないことの検証
					if _, exist := isuIDSet[condition.IsuID]; exist {
						return errorMismatch(res, "同じ ISU のコンディションが複数登録されています")
					}
					isuIDSet[condition.IsuID] = struct{}{}

					// 該当 condition が新規のものである場合はキャッシュを更新
					if !viewer.ConditionAlreadyVerified(condition.IsuID, condition.Timestamp) {
						// 該当 condition が以前のものよりも昔の timestamp で無いことの検証
						if !viewer.ConditionIsUpdated(condition.IsuID, condition.Timestamp) {
							return errorMismatch(res, "以前の取得結果よりも古いタイムスタンプのコンディションが返されています")
						}
						viewer.SetVerifiedCondition(condition.IsuID, condition.Timestamp)
					}

					return nil
				}(); err != nil {
					return err
				}
			}
		}
	}
	// characterSet の検証
	if !characterSet.IsFull() {
		return errorInvalid(res, "全ての性格のトレンドが取得できていません")
	}

	// 初期データのISUが全てあることを確認
	if len(isuIDSet) != s.LenOfIsuFromId() {
		return errorInvalid(res, "ISU の個数が一致しません")
	}
	return nil
}

//分以下を切り捨て、一時間単位にする関数
func truncateAfterHours(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func verifyPrepareGraph(res *http.Response, targetUser *model.User, targetIsuUUID string,
	getGraphReq *service.GetGraphRequest,
	getGraphResp service.GraphResponse) error {

	// graphResp の配列は必ず 24 つ (24時間分) である
	if len(getGraphResp) != 24 {
		return errorInvalid(res, "要素数が正しくありません")
	}

	reqDate := time.Unix(getGraphReq.Date, 0)
	startDate := truncateAfterHours(reqDate).Unix()
	endDate := truncateAfterHours(reqDate.Add(24 * time.Hour)).Unix()

	var lastStartAt int64
	// getGraphResp を逆順 (timestamp が新しい順) にloop
	for idxGraphResp := len(getGraphResp) - 1; idxGraphResp >= 0; idxGraphResp-- {
		graphOne := getGraphResp[idxGraphResp]

		// getGraphResp の要素が古い順に連続して並んでいることの検証
		if idxGraphResp != len(getGraphResp)-1 && !(graphOne.EndAt == lastStartAt) {
			return errorInvalid(res, "整列順が正しくありません")
		}
		lastStartAt = graphOne.StartAt

		// graphのデータが指定日内のものか検証
		if graphOne.StartAt < startDate || endDate < graphOne.EndAt {
			return errorInvalid(res, "グラフの日付が間違っています")
		}

		targetIsu := targetUser.IsuListByID[targetIsuUUID]
		var conditionsBaseOfScore []*model.IsuCondition

		if err := func() error {
			// isu.Condition の read lock を取る
			targetIsu.CondMutex.RLock()
			defer targetIsu.CondMutex.RUnlock()

			// 特定の ISU における expected な conditions を新しい順に取得するイテレータを生成
			filter := model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical
			baseIter := targetIsu.Conditions.LowerBound(filter, graphOne.EndAt)

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

				// ここだけPostConditionが保証できてないLoad時のチェックと異なり全て存在する前提でチェックする
				// expectedの内容がgraphOne.ConditionTimestamps[*]に必ず存在することの検証
				expected := baseIter.Prev()
				if expected != nil && expected.TimestampUnix == timestamp {
					// graphOne.ConditionTimestamps[n] から condition を取得
					conditionsBaseOfScore = append(conditionsBaseOfScore, expected)
				} else {
					return errorMismatch(res, "GraphのTimestampデータが正しくありません")
				}
			}

			expected := baseIter.Prev()
			if expected != nil && graphOne.StartAt <= expected.TimestampUnix {
				return errorMismatch(res, "初期データが欠損しています")
			}

			return nil
		}(); err != nil {
			return err
		}

		// conditionsBaseOfScore と graphOne.Data のどちらか一方が空のときはエラー
		if (len(conditionsBaseOfScore) == 0 || graphOne.Data == nil) && !(len(conditionsBaseOfScore) == 0 && graphOne.Data == nil) {
			return errorMismatch(res, "グラフの data と condition_timestamps のどちらか一方のみが null、あるいは空です。")
		}

		// actual の data が空の場合 verify skip
		if len(conditionsBaseOfScore) == 0 || graphOne.Data == nil {
			continue
		}

		// conditionsBaseOfScore から組み立てた data が actual と等値であることの検証
		expectedGraph := model.NewGraph(conditionsBaseOfScore)
		if !expectedGraph.Match(
			graphOne.Data.Score,
			graphOne.Data.Percentage.Sitting,
			graphOne.Data.Percentage.IsBroken,
			graphOne.Data.Percentage.IsDirty,
			graphOne.Data.Percentage.IsOverweight,
		) {
			return errorMismatch(res, "グラフのデータが正しくありません")
		}
	}

	return nil
}

func verifyPrepareIsuList(res *http.Response, expectedReverse []*model.Isu, isuList []*service.Isu) []error {
	var errs []error
	length := len(expectedReverse)
	if length != len(isuList) {
		errs = append(errs, errorMismatch(res, "椅子の数が異なります"))
		return errs
	}
	for i, isu := range isuList {
		expected := expectedReverse[length-1-i]

		// isu の検証 (jia_isu_uuid, id, character, name)
		err := verifyIsu(res, expected, isu)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		//LatestIsuCondition
		func() {
			expected.CondMutex.RLock()
			defer expected.CondMutex.RUnlock()

			baseIter := expected.Conditions.End(model.ConditionLevelInfo | model.ConditionLevelWarning | model.ConditionLevelCritical)
			expectedCondition := baseIter.Prev()

			// isu.latest_isu_condition が nil &&  前回の latestIsuCondition の timestamp が初期値ならば
			// この ISU は conditionを未送信なのでOK
			if isu.LatestIsuCondition == nil && expectedCondition == nil {
				return
			}

			if isu.LatestIsuCondition == nil {
				errs = append(errs, errorMismatch(res, "%d番目の椅子 (JIA_ISU_UUID=%s) の情報が異なります: 登録されている時刻のデータが返されていません", i+1, isu.JIAIsuUUID))
				return
			}
			if expectedCondition == nil {
				errs = append(errs, errorMismatch(res, "%d番目の椅子 (JIA_ISU_UUID=%s) の情報が異なります: 登録されていない時刻のデータが返されました", i+1, isu.JIAIsuUUID))
				return
			}

			// conditionの検証
			if !(expectedCondition.TimestampUnix == isu.LatestIsuCondition.Timestamp &&
				expected.JIAIsuUUID == isu.LatestIsuCondition.JIAIsuUUID &&
				expected.Name == isu.LatestIsuCondition.IsuName &&
				expectedCondition.IsSitting == isu.LatestIsuCondition.IsSitting &&
				expectedCondition.ConditionString() == isu.LatestIsuCondition.Condition &&
				expectedCondition.ConditionLevel.Equal(isu.LatestIsuCondition.ConditionLevel) &&
				expectedCondition.Message == isu.LatestIsuCondition.Message) {
				errs = append(errs, errorMismatch(res, "%d番目の椅子 (JIA_ISU_UUID=%s) の情報が異なります: latest_isu_conditionの内容が不正です", i+1, isu.JIAIsuUUID))
			}
		}()
	}
	return errs
}

func verifyMe(userID string, hres *http.Response, me *service.GetMeResponse) error {
	if me.JIAUserID != userID {
		return errorInvalid(hres, "ログインユーザと一致しません。")
	}
	return nil
}
