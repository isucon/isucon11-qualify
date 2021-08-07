package scenario

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

// scenario.go
// シナリオ構造体とそのメンバ関数
// および、全ステップで使うシナリオ関数

type Scenario struct {
	// TODO: シナリオ実行に必要なフィールドを書く

	BaseURL                  string        // ベンチ対象 Web アプリの URL
	UseTLS                   bool          // https で接続するかどうか
	NoLoad                   bool          // Load(ベンチ負荷)を強要しない
	LoadTimeout              time.Duration //Loadのcontextの時間
	realTimeLoadFinishedAt   time.Time     //Loadのcontext終了時間
	realTimePrepareStartedAt time.Time     //Prepareの開始時間
	virtualTimeStart         time.Time
	virtualTimeMulti         time.Duration //時間が何倍速になっているか
	jiaServiceURL            *url.URL

	// POST /initialize の猶予時間
	initializeTimeout time.Duration
	prepareTimeout    time.Duration // prepareのTimeoutデフォルト設定

	// 競技者の実装言語
	Language string

	loadWaitGroup sync.WaitGroup
	jiaCancel     context.CancelFunc

	//内部状態
	normalUsersMtx sync.Mutex
	normalUsers    []*model.User

	viewerMtx sync.Mutex
	viewers   []*model.Viewer

	// GET /api/trend にて isuID から isu を取得するのに利用
	isuFromID      map[int]*model.Isu
	isuFromIDMutex sync.RWMutex
}

var (
// MEMO: IsuFromID は NewIsu() 内でのみ書き込まれる append only な map
)

func NewScenario(jiaServiceURL *url.URL, loadTimeout time.Duration) (*Scenario, error) {
	return &Scenario{
		// TODO: シナリオを初期化する
		//realTimeStart: time.Now()
		LoadTimeout:       loadTimeout,
		virtualTimeStart:  random.BaseTime, //初期データ生成時のベースタイムと合わせるために当パッケージの値を利用
		virtualTimeMulti:  30000,           //5分=300秒に一回 => 1秒に100回
		jiaServiceURL:     jiaServiceURL,
		initializeTimeout: 20 * time.Second,
		prepareTimeout:    3 * time.Second,
		normalUsers:       []*model.User{},
		isuFromID:         make(map[int]*model.Isu, 8192),
	}, nil
}

func (s *Scenario) WithInitializeTimeout(t time.Duration) *Scenario {
	s.initializeTimeout = t
	return s
}

func (s *Scenario) NewAgent(opts ...agent.AgentOption) (*agent.Agent, error) {
	opts = append(opts, agent.WithBaseURL(s.BaseURL))
	return agent.NewAgent(opts...)
}

func (s *Scenario) ToVirtualTime(realTime time.Time) time.Time {
	return s.virtualTimeStart.Add(realTime.Sub(s.realTimePrepareStartedAt) * s.virtualTimeMulti)
}

//load用
//通常ユーザーのシナリオ Goroutineを追加する
func (s *Scenario) AddNormalUser(ctx context.Context, step *isucandar.BenchmarkStep, count int) {
	if count <= 0 {
		return
	}
	s.loadWaitGroup.Add(count)
	for i := 0; i < count; i++ {
		go func(ctx context.Context, step *isucandar.BenchmarkStep) {
			defer s.loadWaitGroup.Done()

			user := s.initNormalUser(ctx, step)
			if user == nil {
				return
			}
			user = s.initNormalUserIsu(ctx, step, user)
			if user == nil {
				return
			}
			defer user.CloseAllIsuStateChan()

			step.AddScore(ScoreNormalUserInitialize)

			s.loadNormalUser(ctx, step, user)
		}(ctx, step)
	}
}

// load 中に name が isucon なユーザーを特別に走らせるようにする
func (s *Scenario) AddIsuconUser(ctx context.Context, step *isucandar.BenchmarkStep) {
	s.loadWaitGroup.Add(1)
	go func(ctx context.Context, step *isucandar.BenchmarkStep) {
		defer s.loadWaitGroup.Done()

		user := s.initIsuconUser(ctx, step)
		if user == nil {
			return
		}
		user = s.initNormalUserIsu(ctx, step, user)
		if user == nil {
			return
		}
		defer user.CloseAllIsuStateChan()

		step.AddScore(ScoreNormalUserInitialize)

		s.loadNormalUser(ctx, step, user)
	}(ctx, step)
}

//load用
//非ログインユーザーのシナリオ Goroutineを追加する
func (s *Scenario) AddViewer(ctx context.Context, step *isucandar.BenchmarkStep, count int) {
	if count <= 0 {
		return
	}
	s.loadWaitGroup.Add(count)
	for i := 0; i < count; i++ {
		go func(ctx context.Context, step *isucandar.BenchmarkStep) {
			defer s.loadWaitGroup.Done()
			s.loadViewer(ctx, step)
		}(ctx, step)
	}
}

//新しい登録済みUserの生成
//失敗したらnilを返す
func (s *Scenario) NewUser(ctx context.Context, step *isucandar.BenchmarkStep, a *agent.Agent, userType model.UserType) *model.User {
	user, err := model.NewRandomUserRaw(userType)
	if err != nil {
		logger.AdminLogger.Panic(err)
		return nil
	}

	//backendにpostする
	//TODO: 確率で失敗してリトライする
	_, errs := authAction(ctx, a, user.UserID)
	for _, err := range errs {
		addErrorWithContext(ctx, step, err)
	}
	if len(errs) > 0 {
		return nil
	}
	user.Agent = a

	return user
}

//　isucon Userの生成
//　失敗したらnilを返す
func (s *Scenario) NewIsuconUser(ctx context.Context, step *isucandar.BenchmarkStep, a *agent.Agent, userType model.UserType) *model.User {
	user, err := model.NewIsuconUserRaw(userType)
	if err != nil {
		logger.AdminLogger.Panic(err)
		return nil
	}

	//backendにpostする
	//TODO: 確率で失敗してリトライする
	_, errs := authAction(ctx, a, user.UserID)
	for _, err := range errs {
		addErrorWithContext(ctx, step, err)
	}
	if len(errs) > 0 {
		return nil
	}
	user.Agent = a

	return user
}

//新しい登録済みISUの生成
//失敗したらnilを返す
func (s *Scenario) NewIsu(ctx context.Context, step *isucandar.BenchmarkStep, owner *model.User, addToUser bool, img []byte, retry bool) *model.Isu {
	isu, streamsForPoster, err := model.NewRandomIsuRaw(owner)
	if err != nil {
		logger.AdminLogger.Panic(err)
		return nil
	}

	//ISU協会にIsu*を登録する必要あり
	RegisterToJiaAPI(isu, streamsForPoster)

	//backendにpostする
	//TODO: 確率で失敗してリトライする
	req := service.PostIsuRequest{
		JIAIsuUUID: isu.JIAIsuUUID,
		IsuName:    isu.Name,
	}
	if img != nil {
		req.Img = img
		isu.SetImage(req.Img)
	}

	if retry {
		res := postIsuInfinityRetry(ctx, owner.Agent, req, step)
		// res == nil => ctx.Done
		if res == nil {
			return nil
		}
	} else {
		_, _, err := postIsuAction(ctx, owner.Agent, req)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return nil
		}
	}

	var res *http.Response
	var isuResponse *service.Isu
	if retry {
		isuResponse, res = getIsuInfinityRetry(ctx, owner.Agent, req.JIAIsuUUID, step)
		// isuResponse == nil => ctx.Done
		if isuResponse == nil {
			return nil
		}
	} else {
		isuResponse, res, err = getIsuIdAction(ctx, owner.Agent, req.JIAIsuUUID)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return nil
		}
	}
	// TODO: これは validate でやるべきなきがする
	if isuResponse.JIAIsuUUID != isu.JIAIsuUUID ||
		isuResponse.Name != isu.Name ||
		isuResponse.Character != isu.Character {
		step.AddError(errorMismatch(res, "レスポンスBodyが正しくありません"))
	}

	// POST isu のレスポンスより ID を取得して isu モデルに代入する
	isu.ID = isuResponse.ID

	// isu.ID から model.TrendCondition を取得できるようにする (GET /trend 用)
	s.UpdateIsuFromID(isu)

	// poster に isu model の初期化終了を伝える
	isu.StreamsForScenario.StateChan <- model.IsuStateChangeNone

	//並列に生成する場合は後でgetにより正しい順番を得て、その順序でaddする。企業ユーザーは並列にaddしないと回らない
	//その場合はaddToUser==falseになる
	if addToUser {
		//戻り値をownerに追加する
		owner.AddIsu(isu)
	}
	//投げた時間を
	isu.PostTime = s.ToVirtualTime(time.Now())

	return isu
}

func addErrorWithContext(ctx context.Context, step *isucandar.BenchmarkStep, err error) {
	select {
	case <-ctx.Done():
		if !failure.IsCode(err, ErrHTTP) {
			step.AddError(err)
		}
	default:
		step.AddError(err)
	}
}

func (s *Scenario) UpdateIsuFromID(isu *model.Isu) {
	s.isuFromIDMutex.Lock()
	defer s.isuFromIDMutex.Unlock()
	s.isuFromID[isu.ID] = isu
}

func (s *Scenario) GetIsuFromID(id int) (*model.Isu, bool) {
	s.isuFromIDMutex.RLock()
	defer s.isuFromIDMutex.RUnlock()
	isu, ok := s.isuFromID[id]
	return isu, ok
}
