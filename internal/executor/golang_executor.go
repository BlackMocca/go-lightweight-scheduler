package executor

import (
	"context"
)

type GolangExecuter struct {
	fn func(ctx context.Context) (interface{}, error)
}

func NewGolangExecuter(fn func(ctx context.Context) (interface{}, error)) GolangExecuter {
	return GolangExecuter{fn: fn}
}

func (g GolangExecuter) Execute(ctx context.Context) (interface{}, error) {
	return g.fn(ctx)
}
