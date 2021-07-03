package scenario

import "github.com/isucon/isucandar/agent"

// scenario.go
// シナリオ構造体とそのメンバ関数

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
