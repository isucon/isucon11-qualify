package scenario

import (
	"context"

	"github.com/isucon/isucandar"
)

func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	if s.NoLoad {
		return nil
	}

	/*
		TODO: 負荷走行後のデータ検証シナリオ
	*/

	return nil
}
