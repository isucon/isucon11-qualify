package scenario

// prepare.go
// シナリオの内、prepareフェーズの処理

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/isucon/isucon11-qualify/bench/model"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	logger.ContestantLogger.Printf("===> PREPARE")
	// keepPostingのuserTimerでctx終了させられてしまうのでprepareでも設定する

	//TODO: 他の得点源
	//TODO: 得点調整
	step.Result().Score.Set(ScoreNormalUserInitialize, 10)
	step.Result().Score.Set(ScoreGraphExcellent, 5)
	step.Result().Score.Set(ScoreGraphGood, 4)
	step.Result().Score.Set(ScoreGraphNormal, 3)
	step.Result().Score.Set(ScoreGraphBad, 2)
	step.Result().Score.Set(ScoreGraphWorst, 1)
	step.Result().Score.Set(ScoreReadInfoCondition, 3)
	step.Result().Score.Set(ScoreReadWarningCondition, 2)
	step.Result().Score.Set(ScoreReadCriticalCondition, 1)

	//初期データの生成
	logger.AdminLogger.Println("start: load initial data")
	s.InitializeData()
	logger.AdminLogger.Println("finish: load initial data")
	s.realTimePrepareStartedAt = time.Now()

	//jiaの起動
	s.loadWaitGroup.Add(1)
	ctxJIA, jiaCancelFunc := context.WithCancel(context.Background())
	s.jiaCancel = jiaCancelFunc
	go func() {
		defer s.loadWaitGroup.Done()
		s.JiaAPIService(ctxJIA, step)
	}()
	jiaWait := time.After(10 * time.Second)

	//initialize
	initializer, err := s.NewAgent(
		agent.WithNoCache(), agent.WithNoCookie(), agent.WithTimeout(s.initializeTimeout),
	)
	if err != nil {
		logger.AdminLogger.Panic(err)
	}
	initializer.Name = "benchmarker-initializer"

	initResponse, errs := initializeAction(ctx, initializer, service.PostInitializeRequest{JIAServiceURL: s.jiaServiceURL.String()})
	for _, err := range errs {
		step.AddError(err)
	}
	if len(errs) > 0 {
		//return ErrScenarioCancel
		step.AddError(failure.NewError(ErrCritical, fmt.Errorf("initializeに失敗しました")))
		return nil
	}

	s.Language = initResponse.Language

	//jia起動待ち TODO: これで本当に良いのか？
	<-jiaWait

	//各エンドポイントのチェック
	err = s.prepareCheckAuth(ctx, step)
	if err != nil {
		return err
	}

	if err := s.prepareCheck(ctx, step); err != nil {
		return failure.NewError(ErrCritical, err)
	}
	errors := step.Result().Errors
	hasErrors := func() bool {
		errors.Wait()
		return len(errors.All()) > 0
	}
	// Prepare step でのエラーはすべて Critical の扱い
	if hasErrors() {
		//return ErrScenarioCancel
		step.AddError(failure.NewError(ErrCritical, fmt.Errorf("アプリケーション互換性チェックに失敗しました")))
		return nil
	}

	s.realTimeLoadFinishedAt = time.Now().Add(s.LoadTimeout)
	return nil
}

//エンドポイント事の単体テスト
func (s *Scenario) prepareCheck(parent context.Context, step *isucandar.BenchmarkStep) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	//ユーザー作成
	guestAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}

	noIsuAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	noIsuUser := s.NewUser(ctx, step, noIsuAgent, model.UserTypeNormal)

	// TODO: 初期データから適当に選択してるが、テスト用に想定どおりの値をもつinitialdataを生成する必要がありそう
	var randomUser *model.User
	for {
		randomUser = s.normalUsers[0]
		if len(randomUser.IsuListOrderByCreatedAt) >= 2 {
			break
		}
	}

	agt, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	_, errs := authAction(ctx, agt, randomUser.UserID)
	for _, err := range errs {
		step.AddError(err)
		return nil
	}
	randomUser.Agent = agt

	s.prepareCheckPostSignout(ctx, step)
	s.prepareCheckGetMe(ctx, randomUser, guestAgent, step)
	s.prepareCheckGetIsuList(ctx, randomUser, noIsuUser, guestAgent, step)
	s.prepareCheckPostIsu(ctx, randomUser, noIsuUser, guestAgent, step)
	s.prepareCheckGetIsu(ctx, randomUser, noIsuUser, guestAgent, step)
	s.prepareCheckGetIsuIcon(ctx, randomUser, noIsuUser, guestAgent, step)
	s.prepareCheckGetIsuGraph(ctx, randomUser, noIsuUser, guestAgent, step)
	s.prepareCheckGetIsuConditions(ctx, randomUser, noIsuUser, guestAgent, step)
	// TODO: 確率で失敗するようになったので一旦prepareCheckを行わないようにする。方針決まり次第消すか復活させるかする
	//s.prepareCheckPostIsuCondition(ctx, loginUser, noIsuUser, guestAgent, step)

	return nil
}

func (s *Scenario) prepareCheckAuth(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {

		agt, err := s.NewAgent()
		if err != nil {
			logger.AdminLogger.Panic(err)
			return
		}
		userID := random.UserName()
		if (index % 10) < authActionErrorNum {
			//各種ログイン失敗ケース
			errs := authActionError(ctx, agt, userID, index%10)
			for _, err := range errs {
				step.AddError(err)
			}
		} else {
			//ログイン成功
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(20))

	if err != nil {
		logger.AdminLogger.Panic(err)
	}

	w.Process(ctx)
	//w.Wait()
	//MEMO: ctx.Done()の場合は、プロセスが終了していない可能性がある。

	//作成済みユーザーへのログイン確認
	agt, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panic(err)
		return nil
	}
	userID := random.UserName()

	_, errs := authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
	}
	agt.ClearCookie()
	//二回目のログイン
	_, errs = authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
	}

	return nil
}

func (s *Scenario) prepareCheckPostSignout(ctx context.Context, step *isucandar.BenchmarkStep) {
	// 正常にサインアウト実行
	agt, err := s.NewAgent()
	if err != nil {
		step.AddError(err)
		return
	}
	userID := random.UserName()
	_, errs := authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
		return
	}

	if err := BrowserAccess(ctx, agt, "/login"); err != nil {
		step.AddError(err)
		return
	}
	_, err = signoutAction(ctx, agt)
	if err != nil {
		step.AddError(err)
		return
	}

	// サインインしてない状態でサインアウト実行
	resBody, res, err := signoutErrorAction(ctx, agt)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

}

func (s *Scenario) prepareCheckGetMe(ctx context.Context, loginUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	// 正常系
	meRes, res, err := getMeAction(ctx, loginUser.Agent)
	if err != nil {
		step.AddError(err)
		return
	}

	if meRes == nil {
		step.AddError(errorInvalid(res, "レスポンス内容が不正です。"))
		return
	}
	if meRes.JIAUserID != loginUser.UserID {
		step.AddError(errorInvalid(res, "ログインユーザと一致しません。"))
		return
	}

	// サインインしてない状態で取得
	resBody, res, err := getMeErrorAction(ctx, guestAgent)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}
}

func (s *Scenario) prepareCheckGetIsuList(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	//ISU一覧の取得 e.GET("/api/isu", getIsuList)
	// check: 椅子未所持の場合は椅子が存在しない
	isuList, res, err := getIsuAction(ctx, noIsuUser.Agent, 1)
	if err != nil {
		step.AddError(err)
		return
	}
	if len(isuList) != 0 {
		step.AddError(errorInvalid(res, "ユーザの所持する椅子の数が一致しません。"))
		return
	}

	// check: 登録したISUがlimit分取得できる
	// TODO: 2個以上の椅子をもたせる
	size := len(loginUser.IsuListOrderByCreatedAt)
	isu2 := loginUser.IsuListOrderByCreatedAt[size-2]
	isu3 := loginUser.IsuListOrderByCreatedAt[size-1]
	isuList, res, err = getIsuAction(ctx, loginUser.Agent, 2)
	if err != nil {
		step.AddError(err)
		return
	}
	//expected
	sIsu2 := isu2.ToService()
	sIsu3 := isu3.ToService()
	expected := []*service.Isu{sIsu3, sIsu2}
	if !reflect.DeepEqual(isuList, expected) {
		step.AddError(errorInvalid(res, "ユーザの所持する椅子の数や順番が一致しません。"))
		return
	}

	// check: サインインしてない状態で取得
	query := url.Values{}
	resBody, res, err := getIsuErrorAction(ctx, guestAgent, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

	// check: limitが不正(0)
	query = url.Values{}
	query.Set("limit", "0")
	resBody, res, err = getIsuErrorAction(ctx, loginUser.Agent, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: limit"); err != nil {
		step.AddError(err)
		return
	}

	// check: limitが不正(文字列)
	query = url.Values{}
	query.Set("limit", "limit")
	resBody, res, err = getIsuErrorAction(ctx, loginUser.Agent, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: limit"); err != nil {
		step.AddError(err)
		return
	}
}

func (s *Scenario) prepareCheckPostIsu(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	//Isuの登録 e.POST("/api/isu", postIsu)
	// check: 椅子の登録が成功する（デフォルト画像）
	isu := s.NewIsu(ctx, step, loginUser, true, nil)
	if isu == nil {
		return
	}

	expected := isu.ToService()
	actual, res, err := getIsuIdAction(ctx, loginUser.Agent, isu.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	if !reflect.DeepEqual(*actual, *expected) {
		step.AddError(errorInvalid(res, "ユーザが所持している椅子が取得できません。"))
		return
	}

	imgByte, res, err := getIsuIconAction(ctx, loginUser.Agent, isu.JIAIsuUUID, false)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusOK); err != nil {
		step.AddError(err)
		return
	}
	data, err := ioutil.ReadFile("./images/default.jpg")
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	expectedImg := md5.Sum(data)
	actualImg := md5.Sum(imgByte)
	if expectedImg != actualImg {
		// TODO: backendの画像変更待ち
		//step.AddError(errorInvalid(res, "期待するISUアイコンと一致しません"))
		//return
	}

	// check: 椅子の登録が成功する（画像あり）
	img, err := ioutil.ReadFile("./images/CIMG8423_resize.jpg")
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	isuImg := &service.IsuImg{
		ImgName: "CIMG8423_resize.jpg",
		Img:     img,
	}
	isuWithImg := s.NewIsu(ctx, step, loginUser, true, isuImg)
	if isuWithImg == nil {
		return
	}

	expected = isuWithImg.ToService()
	actual, res, err = getIsuIdAction(ctx, loginUser.Agent, isuWithImg.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	if !reflect.DeepEqual(*actual, *expected) {
		step.AddError(errorInvalid(res, "ユーザが所持している椅子が取得できません。"))
		return
	}

	imgByte, res, err = getIsuIconAction(ctx, loginUser.Agent, isuWithImg.JIAIsuUUID, false)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusOK); err != nil {
		step.AddError(err)
		return
	}
	expectedImg = md5.Sum(img)
	actualImg = md5.Sum(imgByte)
	if expectedImg != actualImg {
		step.AddError(errorInvalid(res, "期待するISUアイコンと一致しません"))
		return
	}

	// check: サインインしてない状態で椅子登録
	req := service.PostIsuRequest{
		JIAIsuUUID: isu.JIAIsuUUID,
		IsuName:    isu.Name,
	}
	resBody, res, err := postIsuErrorAction(ctx, guestAgent, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

	// check: 登録済みのisuをactivate
	resBody, res, err = postIsuErrorAction(ctx, loginUser.Agent, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusConflict); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "duplicated: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: 他ユーザがactivate済みのisuをactivate
	resBody, res, err = postIsuErrorAction(ctx, noIsuUser.Agent, req)
	if err != nil {
		step.AddError(err)
		return
	}
	// もともとjia serviceでforbiddenで返されてたけど、バックエンド実装変更でconflictに
	if err := verifyStatusCode(res, http.StatusConflict); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "duplicated: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: 存在しない椅子を登録
	req = service.PostIsuRequest{
		JIAIsuUUID: "jiaisuuuid",
		IsuName:    "isuname",
	}
	resBody, res, err = postIsuErrorAction(ctx, loginUser.Agent, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "JIAService returned error"); err != nil {
		step.AddError(err)
		return
	}
}

func (s *Scenario) prepareCheckGetIsu(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {

	//Isuの詳細情報取得 e.GET("/api/isu/:jia_isu_uuid", getIsu)
	// check: 正常系
	for _, isu := range loginUser.IsuListOrderByCreatedAt {
		if err := BrowserAccess(ctx, loginUser.Agent, "/isu/"+isu.JIAIsuUUID); err != nil {
			step.AddError(err)
			return
		}
		expected := isu.ToService()
		resIsu, res, err := getIsuIdAction(ctx, loginUser.Agent, isu.JIAIsuUUID)
		if err != nil {
			step.AddError(err)
			return
		}
		if !reflect.DeepEqual(*resIsu, *expected) {
			step.AddError(errorInvalid(res, "ユーザが所持している椅子が取得できません。"))
			return
		}
	}

	isu := loginUser.IsuListOrderByCreatedAt[0]
	// check: 未ログイン状態
	resBody, res, err := getIsuIdErrorAction(ctx, guestAgent, isu.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

	// check: 他ユーザの椅子に対するリクエスト
	resBody, res, err = getIsuIdErrorAction(ctx, noIsuUser.Agent, isu.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: 存在しない椅子を取得
	resBody, res, err = getIsuIdErrorAction(ctx, loginUser.Agent, "jiaisuuuid")
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

}

func (s *Scenario) prepareCheckGetIsuIcon(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	// check: ISUのアイコン取得 e.GET("/api/isu/:jia_isu_uuid/icon", getIsuIcon)
	//- 正常系（初回はnot modified許可しない）
	for _, isu := range loginUser.IsuListOrderByCreatedAt {
		imgByte, res, err := getIsuIconAction(ctx, loginUser.Agent, isu.JIAIsuUUID, false)
		if err != nil {
			step.AddError(err)
			return
		}
		if err := verifyStatusCode(res, http.StatusOK); err != nil {
			step.AddError(err)
			return
		}
		expected := isu.ImageHash
		actual := md5.Sum(imgByte)
		if expected != actual {
			step.AddError(errorInvalid(res, "期待するISUアイコンと一致しません"))
			return
		}

		imgByte, res, err = getIsuIconAction(ctx, loginUser.Agent, isu.JIAIsuUUID, true)
		if err != nil {
			step.AddError(err)
			return
		}
		actual = md5.Sum(imgByte)
		if expected != actual {
			step.AddError(errorInvalid(res, "期待するISUアイコンと一致しません"))
			return
		}
	}

	isu := loginUser.IsuListOrderByCreatedAt[0]
	// check: 未ログイン状態
	resBody, res, err := getIsuIconErrorAction(ctx, guestAgent, isu.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

	// check: 他ユーザの椅子画像取得
	//  - nginxキャッシュで他ユーザが見れたらダメ(cache OKにするならcache時間の検討必要そう
	resBody, res, err = getIsuIconErrorAction(ctx, noIsuUser.Agent, isu.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: 登録されていない椅子に対するリクエスト
	resBody, res, err = getIsuIconErrorAction(ctx, loginUser.Agent, "jiaisuuuid")
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

}

func (s *Scenario) prepareCheckGetIsuGraph(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	//ISUグラフの取得 e.GET("/api/isu/:jia_isu_uuid/graph", getIsuGraph)
	// check: 正常系
	for _, isu := range loginUser.IsuListOrderByCreatedAt {
		// TODO: 2日分くらいconditionある初期データ作る
		lastCond := isu.Conditions.Back()
		// prepare中に追加したISUはconditionが無いためチェックしない
		if lastCond == nil {
			continue
		}
		graph, res, err := getIsuGraphAction(ctx, loginUser.Agent, isu.JIAIsuUUID, service.GetGraphRequest{Date: lastCond.TimestampUnix})
		if err != nil {
			step.AddError(err)
			return
		}
		if err := verifyStatusCode(res, http.StatusOK); err != nil {
			step.AddError(err)
			return
		}
		// graphの検証
		if err := verifyPrepareGraph(res, loginUser, isu.JIAIsuUUID, graph); err != nil {
			step.AddError(err)
			return
		}
	}

	// check: 未ログイン状態
	isu := loginUser.IsuListOrderByCreatedAt[0]
	query := url.Values{}
	query.Set("datetime", strconv.FormatInt(time.Now().Unix(), 10))
	resBody, res, err := getIsuGraphErrorAction(ctx, guestAgent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

	// check: dateパラメータ不足
	query = url.Values{}
	resBody, res, err = getIsuGraphErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "missing: datetime"); err != nil {
		step.AddError(err)
		return
	}

	// check: dateパラメータのフォーマット違反
	query = url.Values{}
	query.Set("datetime", "datetime")
	resBody, res, err = getIsuGraphErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: datetime"); err != nil {
		step.AddError(err)
		return
	}

	// check: 他ユーザの椅子に対するリクエスト
	query = url.Values{}
	query.Set("datetime", strconv.FormatInt(time.Now().Unix(), 10))
	resBody, res, err = getIsuGraphErrorAction(ctx, noIsuUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: 登録されていない椅子に対するリクエスト
	query = url.Values{}
	query.Set("datetime", strconv.FormatInt(time.Now().Unix(), 10))
	resBody, res, err = getIsuGraphErrorAction(ctx, loginUser.Agent, "jiaisuuuid", query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}
}

// TODO: 一部実装途中
func (s *Scenario) prepareCheckGetIsuConditions(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	lastTime := loginUser.IsuListOrderByCreatedAt[0].Conditions.Back().TimestampUnix
	//ISUコンディションの取得 e.GET("/api/condition/:jia_isu_uuid", getIsuConditions)
	//- 正常系
	//	- option無し
	for jiaIsuUUID, isu := range loginUser.IsuListByID {
		endTime := lastTime
		lastCond := isu.Conditions.Back()
		if lastCond != nil {
			endTime = lastCond.TimestampUnix
		}
		req := service.GetIsuConditionRequest{
			StartTime:      nil,
			EndTime:  endTime,
			ConditionLevel: "info,warning,critical",
			Limit:          nil,
		}
		conditionsTmp, res, err := getIsuConditionAction(ctx, loginUser.Agent, jiaIsuUUID, req)
		if err != nil {
			step.AddError(err)
			return
		}
		//検証
		err = verifyPrepareIsuConditions(res, loginUser, jiaIsuUUID, &req, conditionsTmp)
		if err != nil {
			step.AddError(err)
			return
		}
	}

	// TODO: オプション検証
	// check: 正常系
	// - condition指定infoのみ
	// - end_time指定を途中の時間で
	// - start_time指定あり
	// - limit指定あり
	limit := 10
	for jiaIsuUUID, isu := range loginUser.IsuListByID {
		endTime := lastTime
		lastCond := isu.Conditions.Back()
		if lastCond != nil {
			endTime = lastCond.TimestampUnix
		}
		req := service.GetIsuConditionRequest{
			StartTime:      nil,
			EndTime:  endTime,
			ConditionLevel: "info",
			Limit:          &limit,
		}
		conditionsTmp, res, err := getIsuConditionAction(ctx, loginUser.Agent, jiaIsuUUID, req)
		if err != nil {
			step.AddError(err)
			return
		}
		//検証
		err = verifyPrepareIsuConditions(res, loginUser, jiaIsuUUID, &req, conditionsTmp)
		if err != nil {
			step.AddError(err)
			return
		}
	}

	// check: 未ログイン状態
	query := url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	query.Set("condition_level", "info,warning,critical")

	// TODO: 初期データにisuを持たないユーザがいたらダメなので、数ユーザは固定ユーザ作ったほうが良い
	isu := loginUser.IsuListOrderByCreatedAt[0]
	resBody, res, err := getIsuConditionErrorAction(ctx, guestAgent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyNotSignedIn(res, resBody); err != nil {
		step.AddError(err)
		return
	}

	// check: end_timeパラメータ不足
	query = url.Values{}
	query.Set("condition_level", "info,warning,critical")

	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	// MEMO: 他と違ってパラメータ不足がXXX is missingではなくbad format扱い
	if err := verifyText(res, resBody, "bad format: end_time"); err != nil {
		step.AddError(err)
		return
	}

	// check: end_time不正パラメータ
	query = url.Values{}
	query.Set("end_time", "end_time")
	query.Set("condition_level", "info,warning,critical")
	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: end_time"); err != nil {
		step.AddError(err)
		return
	}

	// check: condition_levelパラメータ不足(空文字含む)
	query = url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "missing: condition_level"); err != nil {
		step.AddError(err)
		return
	}

	// check: start_timeフォーマット違反
	query = url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	query.Set("condition_level", "info,warning,critical")
	query.Set("start_time", "start_time")
	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: start_time"); err != nil {
		step.AddError(err)
		return
	}

	// check: limitフォーマット違反
	query = url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	query.Set("condition_level", "info,warning,critical")
	query.Set("limit", "-1")
	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: limit"); err != nil {
		step.AddError(err)
		return
	}

	// check: limitフォーマット違反2
	query = url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	query.Set("condition_level", "info,warning,critical")
	query.Set("limit", "limit")
	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad format: limit"); err != nil {
		step.AddError(err)
		return
	}

	// check: 他ユーザの椅子に対するリクエスト
	query = url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	query.Set("condition_level", "info,warning,critical")
	resBody, res, err = getIsuConditionErrorAction(ctx, noIsuUser.Agent, isu.JIAIsuUUID, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: 登録されていない椅子に対するリクエスト
	query = url.Values{}
	query.Set("end_time", strconv.FormatInt(lastTime, 10))
	query.Set("condition_level", "info,warning,critical")
	resBody, res, err = getIsuConditionErrorAction(ctx, loginUser.Agent, "jiaisuuuid", query)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}
}

func (s *Scenario) prepareCheckPostIsuCondition(ctx context.Context, loginUser *model.User, noIsuUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	// ISUからのcondition送信 e.POST("/api/isu/:jia_isu_uuid/condition", postIsuCondition)
	// - 正常系
	isu := s.NewIsu(ctx, step, loginUser, true, nil)
	if isu == nil {
		return
	}

	// 通常のisu condition送信とかぶらないように未来の日付にしてる
	// TODO: ここは時間表現ではなく、prepare中はkeepPostingさせないなどして制御するか、keepPostingした上で厳し目チェックに
	baseTime := time.Date(2022, 7, 1, 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60))
	var conditionsReq []service.PostIsuConditionRequest
	var expected []*service.GetIsuConditionResponse
	expected = append(expected, &service.GetIsuConditionResponse{
		JIAIsuUUID:     isu.JIAIsuUUID,
		IsuName:        isu.Name,
		Timestamp:      baseTime.Add(10 * time.Minute).Unix(),
		IsSitting:      true,
		Condition:      fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", true, true, true),
		ConditionLevel: "critical",
		Message:        "助けてください",
	})
	expected = append(expected, &service.GetIsuConditionResponse{
		JIAIsuUUID:     isu.JIAIsuUUID,
		IsuName:        isu.Name,
		Timestamp:      baseTime.Add(5 * time.Minute).Unix(),
		IsSitting:      true,
		Condition:      fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", false, true, false),
		ConditionLevel: "warning",
		Message:        "重たいです",
	})
	expected = append(expected, &service.GetIsuConditionResponse{
		JIAIsuUUID:     isu.JIAIsuUUID,
		IsuName:        isu.Name,
		Timestamp:      baseTime.Unix(),
		IsSitting:      true,
		Condition:      fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", false, false, false),
		ConditionLevel: "info",
		Message:        "おはようございます",
	})

	for _, condRes := range expected {
		conditionsReq = append(conditionsReq, service.PostIsuConditionRequest{
			IsSitting: condRes.IsSitting,
			Condition: condRes.Condition,
			Message:   condRes.Message,
			Timestamp: condRes.Timestamp,
		})
	}
	targetURL := fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], isu.JIAIsuUUID)
	httpClient := http.Client{}
	httpClient.Timeout = agent.DefaultRequestTimeout + 5*time.Second
	res, err := postIsuConditionAction(httpClient, targetURL, &conditionsReq)
	if err != nil {
		step.AddError(err)
		return
	}

	limit := len(expected)
	getReq := service.GetIsuConditionRequest{
		StartTime:      nil,
		EndTime:  baseTime.Add(11 * time.Minute).Unix(),
		ConditionLevel: "info,warning,critical",
		Limit:          &limit,
	}
	conditionsRes, res, err := getIsuConditionAction(ctx, loginUser.Agent, isu.JIAIsuUUID, getReq)
	if len(conditionsRes) != len(expected) {
		step.AddError(errorInvalid(res, "condition数が一致しません。"))
	}
	if !reflect.DeepEqual(conditionsRes, expected) {
		step.AddError(errorInvalid(res, "conditionが一致しません。"))
	}

	// check: conditionかぶりのエラー（かぶらなくなったので削除）

	// check: jia_isu_uuidが空文字
	req := []map[string]interface{}{}
	req = append(req, map[string]interface{}{
		"is_sitting": true,
		"condition":  fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", true, true, true),
		"message":    "message",
		"timestamp":  time.Now().Unix()})

	// check: 存在しないjia_isu_uuid
	targetURL = fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], "jiaisuuuid")
	resBody, res, err := postIsuConditionErrorAction(httpClient, targetURL, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusNotFound); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "not found: isu"); err != nil {
		step.AddError(err)
		return
	}

	// check: conditionフォーマットの不正(cond.Timestampが文字列)
	req = []map[string]interface{}{}
	req = append(req, map[string]interface{}{
		"is_sitting": true,
		"condition":  fmt.Sprintf("is_dirty=%v,is_overweight=%v,is_broken=%v", true, true, true),
		"message":    "message",
		"timestamp":  "hoge",
	})

	targetURL = fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], isu.JIAIsuUUID)
	resBody, res, err = postIsuConditionErrorAction(httpClient, targetURL, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad request body"); err != nil {
		step.AddError(err)
		return
	}

	//check: conditionフォーマットの不正(condition空文字)
	req = []map[string]interface{}{}
	req = append(req, map[string]interface{}{
		"is_sitting": true,
		"condition":  "",
		"message":    "message",
		"timestamp":  time.Now().Unix(),
	})
	targetURL = fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], isu.JIAIsuUUID)
	resBody, res, err = postIsuConditionErrorAction(httpClient, targetURL, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad request body"); err != nil {
		step.AddError(err)
		return
	}

	// check: conditionフォーマットの不正
	req = []map[string]interface{}{}
	req = append(req, map[string]interface{}{
		"is_sitting": true,
		"condition":  fmt.Sprintf("is_dirty=%v", "hoge"),
		"message":    "message",
		"timestamp":  time.Now().Unix(),
	})

	targetURL = fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], isu.JIAIsuUUID)
	resBody, res, err = postIsuConditionErrorAction(httpClient, targetURL, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad request body"); err != nil {
		step.AddError(err)
		return
	}

	// check: is_sittingフォーマットの不正
	req = []map[string]interface{}{}
	req = append(req, map[string]interface{}{
		"is_sitting": "hoge",
		"condition":  fmt.Sprintf("is_dirty=%v", "hoge"),
		"message":    "message",
		"timestamp":  time.Now().Unix(),
	})

	targetURL = fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], isu.JIAIsuUUID)
	resBody, res, err = postIsuConditionErrorAction(httpClient, targetURL, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad request body"); err != nil {
		step.AddError(err)
		return
	}

	// check: reqが空
	req = []map[string]interface{}{}
	targetURL = fmt.Sprintf("%s/api/condition/%s", isuTargetBaseUrl[isu.JIAIsuUUID], isu.JIAIsuUUID)
	resBody, res, err = postIsuConditionErrorAction(httpClient, targetURL, req)
	if err != nil {
		step.AddError(err)
		return
	}
	if err := verifyStatusCode(res, http.StatusBadRequest); err != nil {
		step.AddError(err)
		return
	}
	if err := verifyText(res, resBody, "bad request body"); err != nil {
		step.AddError(err)
		return
	}
}
