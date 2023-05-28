package task

import "context"

type Execution interface {
	GetName() string
	Call(ctx context.Context) error
}
