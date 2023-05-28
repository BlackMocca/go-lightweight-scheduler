package executor

import "context"

type Execution interface {
	Execute(ctx context.Context) error
}
