package engine

import "fmt"

var (
	// 表示工作流引擎执行被中断，可能已结束
	ErrTerminated = fmt.Errorf("workflow engine: execution was terminated")
	// 表示工作流引擎在执行子节点时发生异常
	ErrBlocksException = fmt.Errorf("workflow engine: error occurred when executing blocks")
)
