package engine

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
)

type tryCatchNodeExecutor struct {
	nodeExecutor
}

func (ne *tryCatchNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	var engine *workflowEngine
	if we, ok := execCtx.engine.(*workflowEngine); !ok {
		panic("unreachable")
	} else {
		engine = we
	}

	execRes := newNodeExecutionResult(execCtx.Node)

	tryErrs := make([]error, 0)
	tryBlocks := lo.Filter(execCtx.Node.Blocks, func(n *Node, _ int) bool { return n.Type == NodeTypeTryBlock })
	for _, node := range tryBlocks {
		select {
		case <-execCtx.ctx.Done():
			return execRes, execCtx.ctx.Err()
		default:
		}

		err := engine.executeNode(execCtx.Clone(), node)
		if err != nil {
			if errors.Is(err, ErrTerminated) {
				return execRes, err
			}
			tryErrs = append(tryErrs, err)
		}
	}

	if len(tryErrs) > 0 {
		catchErrs := make([]error, 0)
		catchBlocks := lo.Filter(execCtx.Node.Blocks, func(n *Node, _ int) bool { return n.Type == NodeTypeCatchBlock })
		for _, node := range catchBlocks {
			select {
			case <-execCtx.ctx.Done():
				return execRes, execCtx.ctx.Err()
			default:
			}

			err := engine.executeNode(execCtx.Clone(), node)
			if err != nil {
				if errors.Is(err, ErrTerminated) {
					return execRes, err
				}
				catchErrs = append(catchErrs, err)
			}
		}

		errs := make([]error, 0)
		errs = append(errs, tryErrs...)
		errs = append(errs, catchErrs...)
		return execRes, fmt.Errorf("%w: %w", ErrBlocksException, errors.Join(errs...))
	}

	return execRes, nil
}

func newTryCatchNodeExecutor() NodeExecutor {
	return &tryCatchNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
	}
}

type tryBlockNodeExecutor struct {
	nodeExecutor
}

func (ne *tryBlockNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	var engine *workflowEngine
	if we, ok := execCtx.engine.(*workflowEngine); !ok {
		panic("unreachable")
	} else {
		engine = we
	}

	execRes := newNodeExecutionResult(execCtx.Node)

	if err := engine.executeBlocks(execCtx.Clone(), execCtx.Node.Blocks); err != nil {
		return execRes, fmt.Errorf("%w: %w", ErrBlocksException, err)
	}

	return execRes, nil
}

func newTryBlockNodeExecutor() NodeExecutor {
	return &tryBlockNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
	}
}

type catchBlockNodeExecutor struct {
	nodeExecutor
}

func (ne *catchBlockNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	var engine *workflowEngine
	if we, ok := execCtx.engine.(*workflowEngine); !ok {
		panic("unreachable")
	} else {
		engine = we
	}

	if err := engine.executeBlocks(execCtx.Clone(), execCtx.Node.Blocks); err != nil {
		return execRes, fmt.Errorf("%w: %w", ErrBlocksException, err)
	}

	return execRes, nil
}

func newCatchBlockNodeExecutor() NodeExecutor {
	return &catchBlockNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
	}
}
