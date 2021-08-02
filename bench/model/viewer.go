package model

import (
	"github.com/isucon/isucandar/agent"
)

//基本的には一つのシナリオ Goroutine が一つの Viewer を占有する
type Viewer struct {
	ErrorCount uint
	ViewedUpdatedCount uint
	Agent *agent.Agent
}

func NewViewer(agent *agent.Agent) Viewer {
	return Viewer{
		ErrorCount: 0,
		ViewedUpdatedCount: 0,
		Agent: agent,
	}
}
