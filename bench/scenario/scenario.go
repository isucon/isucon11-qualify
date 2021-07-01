package scenario

import (
	"sync"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon11-qualify/bench/model"
)

// scenario.go
// シナリオ構造体とそのメンバ関数
// および、全ステップで使うシナリオ関数

type Scenario struct {
	// TODO: シナリオ実行に必要なフィールドを書く

	BaseURL string // ベンチ対象 Web アプリの URL
	UseTLS  bool   // https で接続するかどうか
	NoLoad  bool   // Load(ベンチ負荷)を強要しない

	// 競技者の実装言語
	Language string
}

func NewScenario() (*Scenario, error) {
	return &Scenario{
		// TODO: シナリオを初期化する
	}, nil
}

func (s *Scenario) NewAgent(opts ...agent.AgentOption) (*agent.Agent, error) {
	opts = append(opts, agent.WithBaseURL(s.BaseURL))
	return agent.NewAgent(opts...)
}

//新しい登録済みISUの生成
func (s *Scenario) NewIsu(step *isucandar.BenchmarkStep, a *agent.Agent, owner *model.User, UserMutex *sync.Mutex) (*model.Isu, error) {
	isu := model.NewRandomIsuRaw(owner)

	//ISU協会にIsu*を登録する必要あり
	RegisterToIsuAPI(isu.PosterChan)

	//backendにpostする
	//isuPostAction() //TODO:
	//TODO: 確率で失敗してリトライする

	//戻り値をownerに追加する必要あり
	if UserMutex != nil {
		UserMutex.Lock()
		defer UserMutex.Unlock()
	}
	owner.AddIsu(isu)

	return isu, nil
}
