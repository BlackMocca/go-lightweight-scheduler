package executor

import (
	"context"
)

type GolangExecuter struct {
	fn func(ctx context.Context) error
}

func NewGolangExecuter(fn func(ctx context.Context) error) GolangExecuter {
	return GolangExecuter{fn: fn}
}

func (g GolangExecuter) Execute(ctx context.Context) error {
	return g.fn(ctx)
}
