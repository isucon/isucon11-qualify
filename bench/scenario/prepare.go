package scenario

import (
	"context"

	"github.com/isucon/isucandar"
)

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	/*
		TODO: 負荷走行前の初期化部分をここに書く(ex: GET /initialize とか)
	*/
	return nil
}
