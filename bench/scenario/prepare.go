package scenario

// prepare.go
// シナリオの内、prepareフェーズの処理

import (
	"context"
	"fmt"
	"github.com/isucon/isucon11-qualify/bench/model"
	"net/http"
	"net/url"
	"reflect"
	"time"

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

	//TODO: 他の得点源
	//TODO: 得点調整
	step.Result().Score.Set(ScoreNormalUserInitialize, 10)
	step.Result().Score.Set(ScoreNormalUserLoop, 10)
	step.Result().Score.Set(ScorePostConditionInfo, 2)
	step.Result().Score.Set(ScorePostConditionWarning, 1)
	step.Result().Score.Set(ScorePostConditionCritical, 0)

	//初期データの生成
	s.InitializeData()
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
		return failure.NewError(ErrCritical, err)
	}
	initializer.Name = "benchmarker-initializer"

	initResponse, errs := initializeAction(ctx, initializer, service.PostInitializeRequest{JIAServiceURL: s.jiaServiceURL.String()})
	for _, err := range errs {
		step.AddError(err)
	}
	if len(errs) > 0 {
		//return ErrScenarioCancel
		return ErrCritical
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
	// Prepare step でのエラーはすべて Critical の扱い
	if len(step.Result().Errors.All()) > 0 {
		//return ErrScenarioCancel
		return ErrCritical
	}

	s.realTimeLoadFinishedAt = time.Now().Add(s.LoadTimeout)
	return nil
}

//エンドポイント事の単体テスト
func (s *Scenario) prepareCheck(parent context.Context, step *isucandar.BenchmarkStep) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	//ユーザー作成
	loginAgent, err := s.NewAgent()
	if err != nil {
		logger.AdminLogger.Panicln(err)
	}
	loginUser := s.NewUser(ctx, step, loginAgent, model.UserTypeNormal)
	guestAgent, err := s.NewAgent()
	if err != nil {
		return err
	}
	// サインアウトの確認
	s.prepareCheckPostSignout(ctx, step)
	s.prepareCheckGetMe(ctx, loginUser, guestAgent, step)
	s.prepareCheckGetIsuList(ctx, loginUser, guestAgent, step)

	// TODO:

	return nil
}

func (s *Scenario) prepareCheckAuth(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {

		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
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
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	//w.Wait()
	//MEMO: ctx.Done()の場合は、プロセスが終了していない可能性がある。

	//作成済みユーザーへのログイン確認
	agt, err := s.NewAgent()
	if err != nil {
		step.AddError(failure.NewError(ErrCritical, err))
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
	logger.AdminLogger.Printf("正常にサインアウト実行")
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
	logger.AdminLogger.Printf("未ログイン状態でサインアウト実行")
	_, err = signoutActionWithoutAuth(ctx, agt)
	if err != nil {
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

	if meRes == nil || meRes.JIAUserID != loginUser.UserID {
		step.AddError(failure.NewError(ErrInvalidResponse, fmt.Errorf("/api/get/meのレスポンス内容が不正です。")))
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

func (s *Scenario) prepareCheckGetIsuList(ctx context.Context, loginUser *model.User, guestAgent *agent.Agent, step *isucandar.BenchmarkStep) {
	//ISU一覧の取得 e.GET("/api/isu", getIsuList)
	// check: 椅子未所持の場合は椅子が存在しない
	query := url.Values{}
	isuList, res, err := getIsuAction(ctx, loginUser.Agent, query)
	if err != nil {
		step.AddError(err)
		return
	}
	if len(isuList) != 0 {
		// TODO: エラーメッセージ直す
		step.AddError(failure.NewError(ErrInvalidResponse, fmt.Errorf("ユーザの所持する椅子や順番が一致しません。")))
		return
	}

	// check: 登録したISUがlimit分取得できる
	isu1 := s.NewIsu(ctx, step, loginUser, true)
	isu2 := s.NewIsu(ctx, step, loginUser, true)
	isu3 := s.NewIsu(ctx, step, loginUser, true)

	query = url.Values{}
	query.Set("limit", "2")
	isuList, res, err = getIsuAction(ctx, loginUser.Agent, query)
	if err != nil {
		step.AddError(err)
		return
	}
	m2s := func(m *model.Isu) service.Isu {
		return service.Isu{
			JIAIsuUUID:   m.JIAIsuUUID,
			Name:         m.Name,
			JIACatalogID: m.JIACatalogID,
			Character:    m.Character,
		}
	}
	//expected
	sIsu1 := m2s(isu1)
	sIsu2 := m2s(isu2)
	sIsu3 := m2s(isu3)
	expected := []*service.Isu{&sIsu3, &sIsu2}
	for _, ex := range expected {
		logger.AdminLogger.Printf("expected: %v", *ex)
	}
	for _, act := range isuList {
		logger.AdminLogger.Printf("actual: %v", *act)
	}
	if !reflect.DeepEqual(isuList, expected) {
		step.AddError(failure.NewError(ErrInvalidResponse, fmt.Errorf("ユーザの所持する椅子や順番が一致しません。")))
		return
	}

	// check: 削除済みの椅子が取得できないことを確認
	// TODO: NewIsuみたいに一括で処理を行う
	_, err = deleteIsuAction(ctx, loginUser.Agent, isu3.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
		return
	}
	loginUser.RemoveIsu(isu3)
	query = url.Values{}
	isuList, res, err = getIsuAction(ctx, loginUser.Agent, query)
	if err != nil {
		step.AddError(err)
		return
	}
	//expected
	expected = []*service.Isu{&sIsu2, &sIsu1}
	for _, ex := range expected {
		logger.AdminLogger.Printf("expected: %v", *ex)
	}
	for _, act := range isuList {
		logger.AdminLogger.Printf("actual: %v", *act)
	}
	if !reflect.DeepEqual(isuList, expected) {
		step.AddError(failure.NewError(ErrInvalidResponse, fmt.Errorf("ユーザの所持する椅子や順番が一致しません。")))
		return
	}

	// check: サインインしてない状態で取得
	query = url.Values{}
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
	if err := verifyText(res, resBody, "invalid value: limit"); err != nil {
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
	if err := verifyText(res, resBody, "invalid value: limit"); err != nil {
		step.AddError(err)
		return
	}
}

func (s *Scenario) prepareCheckPostIsu(ctx context.Context, step *isucandar.BenchmarkStep) {
	//Isuの登録 e.POST("/api/isu", postIsu)
	//- 正常系
	//- 未ログイン状態
	//- 不正なrequestパラメータ
	//- 登録済みのisuをactivate
	//- delete後のisuをactivate
	//- JIAapiのエラー（種類いくつか）
	//Isuの画像登録
	//- 正常系
	//- 未ログイン状態
	//- image画像不足（エラーメッセージがおかしそう）
	//- png以外の画像を弾く
	//- jpgの破損画像渡す
	//- 本人以外の椅子画像更新
	//- 未登録の椅子画像更新
	//- 削除済みの椅子画像更新

}
func (s *Scenario) prepareCheckGetIsu(ctx context.Context, step *isucandar.BenchmarkStep) {

	//Isuの詳細情報取得 e.GET("/api/isu/:jia_isu_uuid", getIsu)
	//- 正常系
	//- 未ログイン状態
	//- 削除済みの椅子に対するリクエスト
	//- 他ユーザの椅子に対するリクエスト
	//- 登録されていない椅子に対するリクエスト

}
func (s *Scenario) prepareCheckDeleteIsu(ctx context.Context, step *isucandar.BenchmarkStep) {
	//ISUの削除 e.DELETE("/api/isu/:jia_isu_uuid", deleteIsu)
	//- 正常系
	//- jiaへのリクエスト成功確認？
	//- 未ログイン状態
	//- 削除済みの椅子に対するリクエスト
	//- 他ユーザの椅子に対するリクエスト
	//- 登録されていない椅子に対するリクエスト

}
func (s *Scenario) prepareCheckGetIsuIcon(ctx context.Context, step *isucandar.BenchmarkStep) {
	//ISUのアイコン取得 e.GET("/api/isu/:jia_isu_uuid/icon", getIsuIcon)
	//- 正常系
	//- 未ログイン状態
	//- 削除済みの椅子に対するリクエスト
	//- cache時間検討必要そう
	//- 他ユーザの椅子に対するリクエスト
	//- nginxキャッシュで他ユーザが見れたらダメ
	//- 登録されていない椅子に対するリクエスト

}
func (s *Scenario) prepareCheckGetIsuGraph(ctx context.Context, step *isucandar.BenchmarkStep) {
	//ISUグラフの取得 e.GET("/api/isu/:jia_isu_uuid/graph", getIsuGraph)
	//- 正常系
	//- 未ログイン状態（不正ログイン）
	//- dateパラメータ不足
	//- dateパラメータのフォーマット違反
	//- 削除済みの椅子に対するリクエスト
	//- 他ユーザの椅子に対するリクエスト
	//- 登録されていない椅子に対するリクエスト

}
func (s *Scenario) prepareCheckGetAllIsuConditions(ctx context.Context, step *isucandar.BenchmarkStep) {
	//ISUコンディションリストの取得 e.GET("/api/condition", getAllIsuConditions)
	//- 正常系
	//- optionあり（組み合わせ）
	//- option無し
	//- userの削除済みでない所持椅子だけか
	//- 未ログイン状態（不正ログイン）
	//- cursor_end_timeパラメータ不足
	//- cursor_end_timeフォーマット違反
	//- cursor_jia_isu_uuidパラメータ不足
	//- condition_levelパラメータ不足(空文字含む)
	//- ,のみはOKなの？
	//- start_timeフォーマット違反
	//- limitフォーマット違反

}
func (s *Scenario) prepareCheckGetIsuConditions(ctx context.Context, step *isucandar.BenchmarkStep) {
	//ISUコンディションの取得 e.GET("/api/condition/:jia_isu_uuid", getIsuConditions)
	//- 正常系
	//- optionあり（組み合わせ）
	//- option無し
	//- 未ログイン状態（不正ログイン）
	//- jia_isu_uuidパラメータ不足(空文字含む)
	//- cursor_end_timeパラメータ不足
	//- cursor_end_timeフォーマット違反
	//- condition_levelパラメータ不足(空文字含む)
	//- ,のみはOKなの？
	//- start_timeフォーマット違反
	//- limitフォーマット違反
	//- 削除済みの椅子に対するリクエスト
	//- 他ユーザの椅子に対するリクエスト
	//- 登録されていない椅子に対するリクエスト
}
func (s *Scenario) prepareCheckPostIsuCondition(ctx context.Context, step *isucandar.BenchmarkStep) {
	// ISUからのcondition送信 e.POST("/api/isu/:jia_isu_uuid/condition", postIsuCondition)
	// - 正常系
	// - 未ログイン状態
	// - parameter不正
	// - 削除済みの椅子に対するリクエスト
	// - 登録されていない椅子に対するリクエスト
	// - conditionフォーマットの不正
	// - conditionかぶりのエラー
}
