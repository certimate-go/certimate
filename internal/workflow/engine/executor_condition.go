package engine

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/samber/lo"
)

type conditionNodeExecutor struct {
	nodeExecutor
}

func (ne *conditionNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	var engine *workflowEngine
	if we, ok := execCtx.engine.(*workflowEngine); !ok {
		panic("unreachable")
	} else {
		engine = we
	}

	execRes := newNodeExecutionResult(execCtx.Node)

	errs := make([]error, 0)
	blocks := lo.Filter(execCtx.Node.Blocks, func(n *Node, _ int) bool { return n.Type == NodeTypeBranchBlock })
	for _, node := range blocks {
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
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return execRes, fmt.Errorf("%w: %w", ErrBlocksException, errors.Join(errs...))
	}

	return execRes, nil
}

func newConditionNodeExecutor() NodeExecutor {
	return &conditionNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
	}
}

type branchBlockNodeExecutor struct {
	nodeExecutor
}

func (ne *branchBlockNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	nodeCfg := execCtx.Node.Data.Config.AsBranchBlock()
	if nodeCfg.Expression == nil {
		ne.logger.Info("enter this branch without any conditions")
	} else {
		variables := lo.Reduce(execCtx.variables.All(), func(acc map[string]map[string]any, state VariableState, _ int) map[string]map[string]any {
			if _, ok := acc[state.Scope]; !ok {
				acc[state.Scope] = make(map[string]any)
			}

			// 这里需要把所有值都转换为字符串形式，因为 Expression.Eval 仅支持字符串类型的值
			acc[state.Scope][state.Key] = state.ValueString()
			return acc
		}, make(map[string]map[string]any))

		rs, err := nodeCfg.Expression.Eval(variables)
		if err != nil {
			ne.logger.Warn(fmt.Sprintf("failed to eval expr: %+v", err))
			return execRes, err
		}

		if rs.Value == false {
			ne.logger.Info("skip this branch, because condition not met")
			return execRes, nil
		} else {
			ne.logger.Info("enter this branch, because condition met")
		}
	}

	if engine, ok := execCtx.engine.(*workflowEngine); !ok {
		panic("unreachable")
	} else {
		if err := engine.executeBlocks(execCtx.Clone(), execCtx.Node.Blocks); err != nil {
			return execRes, fmt.Errorf("%w: %w", ErrBlocksException, err)
		}
	}

	return execRes, nil
}

func newBranchBlockNodeExecutor() NodeExecutor {
	return &branchBlockNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
	}
}
