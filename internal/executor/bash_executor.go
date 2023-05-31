package executor

import (
	"context"
	"fmt"
	"os/exec"
)

type BashExecutor struct {
	cmd        string
	showResult bool
}

func NewBashExecutor(cmd string, showResult bool) Execution {
	return &BashExecutor{
		cmd:        cmd,
		showResult: showResult,
	}
}

func (b BashExecutor) GetName() string {
	return "BashExecutor"
}

func (b BashExecutor) Execute(ctx context.Context) (interface{}, error) {
	cmd := exec.Command("bash", "-c", b.cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if b.showResult {
		fmt.Println(cmd.String())
		fmt.Println(string(output))
	}
	return string(output), nil
}
