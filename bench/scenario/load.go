package scenario

import (
	"context"

	"github.com/isucon/isucandar"
)

func (s *Scenario) Load(parent context.Context, step *isucandar.BenchmarkStep) error {
	if s.NoLoad {
		return nil
	}

	/*
		TODO: 実際の負荷走行シナリオ
	*/

	return nil
}
