package executor

import "context"

type BashExecutor struct {
	cmd  string
	args []string
}

func NewBashExecutor(cmd string, args []string) Execution {
	return &BashExecutor{
		cmd:  cmd,
		args: args,
	}
}

func (b BashExecutor) Execute(ctx context.Context) error {
	return nil
}
