package scenario

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/random/useragent"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/random"
	"github.com/isucon/isucon11-qualify/bench/service"
)

// scenario.go
// シナリオ構造体とそのメンバ関数
// および、全ステップで使うシナリオ関数

type Scenario struct {
	BaseURL                  string        // ベンチ対象 Web アプリの URL
	UseTLS                   bool          // ベンチ対象 Web アプリが HTTPS で動いているかどうか (本番時ture/CI時false)
	NoLoad                   bool          // Load(ベンチ負荷)を強要しない
	LoadTimeout              time.Duration //Loadのcontextの時間
	realTimePrepareStartedAt time.Time     //Prepareの開始時間
	virtualTimeStart         time.Time
	virtualTimeMulti         time.Duration //時間が何倍速になっているか
	jiaServiceURL            *url.URL

	// POST /initialize の猶予時間
	initializeTimeout time.Duration
	prepareTimeout    time.Duration // prepareのTimeoutデフォルト設定

	// 競技者の実装言語
	Language string

	loadWaitGroup   sync.WaitGroup
	JiaPosterCancel context.CancelFunc

	// IPアドレスと FQDN の対応付け
	mapIPAddrToFqdn map[string]string
	mapFqdnToIPAddr map[string]string

	//prepare check用のユーザー
	noIsuUser *model.User

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
		LoadTimeout:       loadTimeout,
		virtualTimeStart:  random.BaseTime, //初期データ生成時のベースタイムと合わせるために当パッケージの値を利用
		virtualTimeMulti:  30000,           //5分=300秒に一回 => 1秒に100回
		jiaServiceURL:     jiaServiceURL,
		initializeTimeout: 20 * time.Second,
		prepareTimeout:    3 * time.Second,
		mapIPAddrToFqdn:   make(map[string]string, 3),
		mapFqdnToIPAddr:   make(map[string]string, 3),
		normalUsers:       make([]*model.User, 0),
		isuFromID:         make(map[int]*model.Isu, 8192),
	}, nil
}

func (s *Scenario) WithInitializeTimeout(t time.Duration) *Scenario {
	s.initializeTimeout = t
	return s
}

func (s *Scenario) separatedTransport() agent.AgentOption {
	return func(a *agent.Agent) error {
		transport := agent.DefaultTransport.Clone()
		transport.MaxIdleConnsPerHost = 100
		a.HttpClient.Transport = transport
		return nil
	}
}

func (s *Scenario) NewAgent(opts ...agent.AgentOption) (*agent.Agent, error) {
	opts = append([]agent.AgentOption{s.separatedTransport()}, opts...)
	opts = append(opts, agent.WithBaseURL(s.BaseURL), agent.WithUserAgent(useragent.UserAgent()))
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
			defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() AddNormalUser")
			s.loadNormalUser(ctx, step, false)
		}(ctx, step)
	}
}

// load 中に name が isucon なユーザーを特別に走らせるようにする
func (s *Scenario) AddIsuconUser(ctx context.Context, step *isucandar.BenchmarkStep) {
	s.loadWaitGroup.Add(1)
	go func(ctx context.Context, step *isucandar.BenchmarkStep) {
		defer s.loadWaitGroup.Done()
		defer logger.AdminLogger.Println("defer s.loadWaitGroup.Done() AddIsuconUser")
		s.loadNormalUser(ctx, step, true)
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
func (s *Scenario) NewUser(ctx context.Context, step *isucandar.BenchmarkStep, a *agent.Agent, userType model.UserType, isIsuconUser bool) *model.User {
	user, err := model.NewRandomUserRaw(userType, isIsuconUser)
	if err != nil {
		logger.AdminLogger.Panic(err)
		return nil
	}
	user.Agent = a

	//backendにpostする
	// 登録済みユーザーは trend に興味がないからリクエストを待たない
	if errs := browserGetLandingPageIgnoreAction(ctx, user); len(errs) != 0 {
		for _, err := range errs {
			addErrorWithContext(ctx, step, err)
		}
	}
	_, errs := authAction(ctx, user, user.UserID)
	for _, err := range errs {
		addErrorWithContext(ctx, step, err)
	}
	if len(errs) > 0 {
		return nil
	}

	// POST /api/auth をしたため GET /api/user/me を叩く
	me, hres, err := getMeAction(ctx, user.Agent)
	if err != nil {
		addErrorWithContext(ctx, step, err)
		// 致命的なエラーではないため return しない
	} else {
		err = verifyMe(user.UserID, hres, me)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			// 致命的なエラーではないため return しない
		}
	}

	return user
}

var newIsuCountForImageMissing int32 = -1 //画像をnilにするかどうかの判定用変数(他用途で使用しないこと)

//新しい登録済みISUの生成
//失敗したらnilを返す
func (s *Scenario) NewIsu(ctx context.Context, step *isucandar.BenchmarkStep, owner *model.User, addToUser bool, retry bool) *model.Isu {
	var image []byte = nil
	//20回に1回はnilでPOST
	if atomic.AddInt32(&newIsuCountForImageMissing, 1)%20 != 0 {
		//画像付きでPOST
		var err error
		image, err = random.Image()
		if err != nil {
			logger.AdminLogger.Panic(err)
		}
	}
	return s.NewIsuWithCustomImg(ctx, step, owner, addToUser, image, retry)
}

//新しい登録済みISUの生成
//失敗したらnilを返す
func (s *Scenario) NewIsuWithCustomImg(ctx context.Context, step *isucandar.BenchmarkStep, owner *model.User, addToUser bool, img []byte, retry bool) *model.Isu {
	isu, streamsForPoster, err := model.NewRandomIsuRaw(owner)
	if err != nil {
		logger.AdminLogger.Panic(err)
		return nil
	}

	//ISU協会にIsu*を登録する必要あり
	RegisterToJiaAPI(isu, streamsForPoster)

	//backendにpostする
	req := service.PostIsuRequest{
		JIAIsuUUID: isu.JIAIsuUUID,
		IsuName:    isu.Name,
	}
	if img != nil {
		req.Img = img
		isu.SetImage(req.Img)
	}

	var res *http.Response
	var isuResponse *service.Isu
	if retry {
		isuResponse, res = postIsuInfinityRetry(ctx, owner.Agent, req, step)
		// res == nil => ctx.Done
		if res == nil {
			return nil
		}
	} else {
		isuResponse, res, err = postIsuAction(ctx, owner.Agent, req)
		if err != nil {
			addErrorWithContext(ctx, step, err)
			return nil
		}
	}
	isuIdUpdated := false
	if res != nil && res.StatusCode == http.StatusConflict {
		// pass
	} else {
		// POST isu のレスポンスより ID を取得して isu モデルに代入する
		isu.ID = isuResponse.ID
		isuIdUpdated = true
		err = verifyIsu(res, isu, isuResponse)
		if err != nil {
			step.AddError(err)
		}
	}

	// var res *http.Response
	// var isuResponse *service.Isu
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
	// GET isu のレスポンスより ID を取得して isu モデルに代入する
	if !isuIdUpdated {
		isu.ID = isuResponse.ID
		isuIdUpdated = true
	}
	err = verifyIsu(res, isu, isuResponse)
	if err != nil {
		step.AddError(err)
	}

	//Icon取得
	icon, res, err := getIsuIconAction(ctx, owner.Agent, isu.JIAIsuUUID)
	if err != nil {
		step.AddError(err)
	} else {
		err = verifyIsuIcon(isu, icon, res.StatusCode)
		if err != nil {
			step.AddError(err)
		}
	}

	// isu.ID から model.TrendCondition を取得できるようにする (GET /trend 用)
	s.UpdateIsuFromID(isu)

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
		return
	default:
		step.AddError(err)
	}
}

// map に対する Setter は main でしか呼ばれないため lock は取らない
func (s *Scenario) SetIPAddrAndFqdn(str ...string) error {
	if len(str)%2 != 0 {
		return fmt.Errorf("invalid arguments")
	}
	for i := 0; i < len(str); i += 2 {
		ipaddr := str[i]
		fqdn := str[i+1]
		s.mapFqdnToIPAddr[fqdn] = ipaddr
		s.mapIPAddrToFqdn[ipaddr] = fqdn
	}
	return nil
}

func (s *Scenario) GetIPAddrFromFqdn(fqdn string) (string, bool) {
	result, ok := s.mapFqdnToIPAddr[fqdn]
	return result, ok
}

func (s *Scenario) GetFqdnFromIPAddr(ipaddr string) (string, bool) {
	result, ok := s.mapIPAddrToFqdn[ipaddr]
	return result, ok
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

func (s *Scenario) LenOfIsuFromId() int {
	s.isuFromIDMutex.RLock()
	defer s.isuFromIDMutex.RUnlock()
	return len(s.isuFromID)
}

func (s *Scenario) GetRandomActivatedIsu(randEngine *rand.Rand) *model.Isu {
	var isu *model.Isu

	s.isuFromIDMutex.RLock()
	defer s.isuFromIDMutex.RUnlock()
	targetCount := randEngine.Intn(len(s.isuFromID))
	for _, isuP := range s.isuFromID {
		if !isuP.IsNoPoster() {
			isu = isuP
		}
		if targetCount <= 0 && isu != nil {
			return isu
		}
		targetCount--
	}
	return isu
}
