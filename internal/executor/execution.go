package executor

import "context"

type Execution interface {
	GetName() string
	Execute(ctx context.Context) (interface{}, error)
}
