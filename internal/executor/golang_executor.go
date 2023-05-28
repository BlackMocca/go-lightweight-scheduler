package executor

import (
	"context"
	"fmt"
)

type GolangExecuter struct {
	fn func(ctx context.Context) error
}

func NewGolangExecuter(fn func(ctx context.Context) error) GolangExecuter {
	return GolangExecuter{fn: fn}
}

func (g GolangExecuter) Execute(ctx context.Context) error {
	fmt.Println("golang Execute")
	return g.fn(ctx)
}
