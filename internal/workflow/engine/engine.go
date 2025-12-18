package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/pkg/logging"
)

type WorkflowExecution struct {
	WorkflowId   string
	WorkflowName string
	RunId        string
	RunTrigger   domain.WorkflowTriggerType
	Graph        *Graph
}

type WorkflowEngine interface {
	Invoke(ctx context.Context, execution WorkflowExecution) error

	OnStart(callback func(ctx context.Context) error)
	OnEnd(callback func(ctx context.Context) error)
	OnError(callback func(ctx context.Context, err error) error)
	OnNodeStart(callback func(ctx context.Context, node *Node) error)
	OnNodeEnd(callback func(ctx context.Context, node *Node, res *NodeExecutionResult) error)
	OnNodeError(callback func(ctx context.Context, node *Node, err error) error)
	OnNodeLogging(callback func(ctx context.Context, node *Node, log logging.Record) error)
}

type workflowEngine struct {
	executors map[NodeType]NodeExecutor

	hooksMtx           sync.RWMutex
	onStartHooks       [](func(ctx context.Context) error)
	onEndHooks         [](func(ctx context.Context) error)
	onErrorHooks       [](func(ctx context.Context, err error) error)
	onNodeStartHooks   [](func(ctx context.Context, node *Node) error)
	onNodeEndHooks     [](func(ctx context.Context, node *Node, res *NodeExecutionResult) error)
	onNodeErrorHooks   [](func(ctx context.Context, node *Node, err error) error)
	onNodeLoggingHooks [](func(ctx context.Context, node *Node, log logging.Record) error)

	wfoutputRepo workflowOutputRepository

	syslog *slog.Logger
}

var _ WorkflowEngine = (*workflowEngine)(nil)

func (we *workflowEngine) Invoke(ctx context.Context, execution WorkflowExecution) error {
	defer func() {
		if r := recover(); r != nil {
			we.fireOnErrorHooks(ctx, fmt.Errorf("workflow engine panic: %v", r))
			we.syslog.Error(fmt.Sprintf("workflow engine panic: %v", r), slog.String("workflowId", execution.WorkflowId), slog.String("runId", execution.RunId))
			slog.Error(fmt.Sprintf("workflow engine panic: %v, stack trace: %s", r, string(debug.Stack())), slog.String("workflowId", execution.WorkflowId), slog.String("runId", execution.RunId))
		}
	}()

	we.fireOnStartHooks(ctx)

	wfIOs := newInOutManager()

	wfVars := newVariableManager()
	wfVars.Set(stateVarKeyWorkflowId, execution.WorkflowId, "string")
	wfVars.Set(stateVarKeyWorkflowName, execution.WorkflowName, "string")
	wfVars.Set(stateVarKeyRunId, execution.RunId, "string")
	wfVars.Set(stateVarKeyRunTrigger, execution.RunTrigger, "string")
	wfVars.Set(stateVarKeyErrorNodeId, "", "string")
	wfVars.Set(stateVarKeyErrorNodeName, "", "string")
	wfVars.Set(stateVarKeyErrorMessage, "", "string")

	wfCtx := (&WorkflowContext{}).
		SetExecutingWorkflow(execution.WorkflowId, execution.RunId, execution.Graph).
		SetEngine(we).
		SetInputsManager(wfIOs).
		SetVariablesManager(wfVars).
		SetContext(ctx)
	if err := we.executeBlocks(wfCtx, execution.Graph.Nodes); err != nil {
		if !errors.Is(err, ErrTerminated) {
			we.fireOnErrorHooks(ctx, err)
			return err
		}
	}

	we.fireOnEndHooks(ctx)

	return nil
}

func (we *workflowEngine) OnStart(callback func(ctx context.Context) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onStartHooks = append(we.onStartHooks, callback)
}

func (we *workflowEngine) OnEnd(callback func(ctx context.Context) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onEndHooks = append(we.onEndHooks, callback)
}

func (we *workflowEngine) OnError(callback func(ctx context.Context, err error) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onErrorHooks = append(we.onErrorHooks, callback)
}

func (we *workflowEngine) OnNodeStart(callback func(ctx context.Context, node *Node) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onNodeStartHooks = append(we.onNodeStartHooks, callback)
}

func (we *workflowEngine) OnNodeEnd(callback func(ctx context.Context, node *Node, res *NodeExecutionResult) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onNodeEndHooks = append(we.onNodeEndHooks, callback)
}

func (we *workflowEngine) OnNodeError(callback func(ctx context.Context, node *Node, err error) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onNodeErrorHooks = append(we.onNodeErrorHooks, callback)
}

func (we *workflowEngine) OnNodeLogging(callback func(ctx context.Context, node *Node, log logging.Record) error) {
	we.hooksMtx.Lock()
	defer we.hooksMtx.Unlock()
	we.onNodeLoggingHooks = append(we.onNodeLoggingHooks, callback)
}

func (we *workflowEngine) executeNode(wfCtx *WorkflowContext, node *Node) error {
	executor, ok := we.executors[node.Type]
	if !ok {
		err := fmt.Errorf("workflow engine: no executor registered for node type: '%s'", node.Type)
		return err
	} else {
		logger := slog.New(logging.NewHookHandler(&logging.HookHandlerOptions{
			Level: slog.LevelDebug,
			WriteFunc: func(ctx context.Context, record logging.Record) error {
				we.fireOnNodeLoggingHooks(ctx, node, record)
				return nil
			},
		}))
		executor.SetLogger(logger)
	}

	wfCtx.variables.SetScoped(node.Id, stateVarKeyNodeId, node.Id, "string")
	wfCtx.variables.SetScoped(node.Id, stateVarKeyNodeName, node.Data.Name, "string")

	// 节点已禁用，直接跳过执行
	if node.Data.Disabled {
		return nil
	}

	we.fireOnNodeStartHooks(wfCtx.ctx, node)

	execCtx := newNodeExecutionContext(wfCtx, node)
	execRes, err := executor.Execute(execCtx)
	if err != nil && !errors.Is(err, ErrTerminated) {
		if !errors.Is(err, ErrBlocksException) {
			wfCtx.variables.Set(stateVarKeyErrorNodeId, node.Id, "string")
			wfCtx.variables.Set(stateVarKeyErrorNodeName, node.Data.Name, "string")
			wfCtx.variables.Set(stateVarKeyErrorMessage, err.Error(), "string")
		}

		we.fireOnNodeErrorHooks(wfCtx.ctx, node, err)
		return err
	}

	we.fireOnNodeEndHooks(wfCtx.ctx, node, execRes)

	if execRes != nil {
		if execRes.Variables != nil {
			for _, variable := range execRes.Variables {
				wfCtx.variables.Add(variable)
			}
		}

		if execRes.Outputs != nil {
			for _, output := range execRes.Outputs {
				wfCtx.inputs.Add(output)
			}
		}

		execOutputs := lo.Filter(execRes.Outputs, func(state InOutState, _ int) bool { return state.Persistent })
		if execRes.outputForced || len(execOutputs) > 0 {
			output := &domain.WorkflowOutput{
				WorkflowId: execCtx.WorkflowId,
				RunId:      execCtx.RunId,
				NodeId:     execCtx.Node.Id,
				NodeConfig: execCtx.Node.Data.Config,
				Succeeded:  true, // 目前恒为 true
			}
			if len(execOutputs) > 0 {
				output.Outputs = lo.Map(execOutputs, func(state InOutState, _ int) *domain.WorkflowOutputEntry {
					return &domain.WorkflowOutputEntry{
						Name:      state.Name,
						Type:      state.Type,
						Value:     state.ValueString(),
						ValueType: state.ValueType,
					}
				})
			}
			if _, err := we.wfoutputRepo.Save(execCtx.ctx, output); err != nil {
				we.syslog.Error("failed to save node output", slog.Any("error", err))
			}
		}

		if execRes.Terminated {
			return ErrTerminated
		}
	}

	if err != nil && errors.Is(err, ErrTerminated) {
		return err
	}

	return nil
}

func (we *workflowEngine) executeBlocks(wfCtx *WorkflowContext, blocks []*Node) error {
	errs := make([]error, 0)

	for _, node := range blocks {
		select {
		case <-wfCtx.ctx.Done():
			return wfCtx.ctx.Err()
		default:
		}

		err := we.executeNode(wfCtx, node)
		if err != nil {
			// 如果当前节点是 TryCatch 节点、且在 CatchBlock 分支中没有 End 节点，
			// 则暂存错误，但继续执行下一个节点，直到当前 Blocks 全部执行完毕。
			if node.Type == NodeTypeTryCatch {
				if !errors.Is(err, ErrTerminated) {
					errs = append(errs, err)
					continue
				}
			}
			return err
		}
	}

	if len(errs) > 0 {
		if len(errs) == 1 {
			return errs[0]
		}
		return errors.Join(errs...)
	}

	return nil
}

func (we *workflowEngine) fireOnStartHooks(ctx context.Context) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onStartHooks {
		if cbErr := cb(ctx); cbErr != nil {
			we.syslog.Error("workflow engine: error in onStart hook", slog.Any("error", cbErr))
		}
	}
}

func (we *workflowEngine) fireOnEndHooks(ctx context.Context) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onEndHooks {
		if cbErr := cb(ctx); cbErr != nil {
			we.syslog.Error("workflow engine: error in onEnd hook", slog.Any("error", cbErr))
		}
	}
}

func (we *workflowEngine) fireOnErrorHooks(ctx context.Context, err error) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onErrorHooks {
		if cbErr := cb(ctx, err); cbErr != nil {
			we.syslog.Error("workflow engine: error in onError hook", slog.Any("error", cbErr))
		}
	}
}

func (we *workflowEngine) fireOnNodeStartHooks(ctx context.Context, node *Node) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onNodeStartHooks {
		if cbErr := cb(ctx, node); cbErr != nil {
			we.syslog.Error("workflow engine: error in onNodeStart hook", slog.Any("error", cbErr))
		}
	}
}

func (we *workflowEngine) fireOnNodeEndHooks(ctx context.Context, node *Node, result *NodeExecutionResult) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onNodeEndHooks {
		if cbErr := cb(ctx, node, result); cbErr != nil {
			we.syslog.Error("workflow engine: error in onNodeEnd hook", slog.Any("error", cbErr))
		}
	}
}

func (we *workflowEngine) fireOnNodeErrorHooks(ctx context.Context, node *Node, err error) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onNodeErrorHooks {
		if cbErr := cb(ctx, node, err); cbErr != nil {
			we.syslog.Error("workflow engine: error in onNodeError hook", slog.Any("error", cbErr))
		}
	}
}

func (we *workflowEngine) fireOnNodeLoggingHooks(ctx context.Context, node *Node, log logging.Record) {
	we.hooksMtx.RLock()
	defer we.hooksMtx.RUnlock()
	for _, cb := range we.onNodeLoggingHooks {
		if cbErr := cb(ctx, node, log); cbErr != nil {
			we.syslog.Error("workflow engine: error in onNodeLogging hook", slog.Any("error", cbErr))
		}
	}
}

func NewWorkflowEngine() WorkflowEngine {
	engine := &workflowEngine{
		executors:    make(map[NodeType]NodeExecutor),
		wfoutputRepo: repository.NewWorkflowOutputRepository(),
		syslog:       app.GetLogger(),
	}
	engine.executors[NodeTypeStart] = newStartNodeExecutor()
	engine.executors[NodeTypeEnd] = newEndNodeExecutor()
	engine.executors[NodeTypeDelay] = newDelayNodeExecutor()
	engine.executors[NodeTypeCondition] = newConditionNodeExecutor()
	engine.executors[NodeTypeBranchBlock] = newBranchBlockNodeExecutor()
	engine.executors[NodeTypeTryCatch] = newTryCatchNodeExecutor()
	engine.executors[NodeTypeTryBlock] = newTryBlockNodeExecutor()
	engine.executors[NodeTypeCatchBlock] = newCatchBlockNodeExecutor()
	engine.executors[NodeTypeBizApply] = newBizApplyNodeExecutor()
	engine.executors[NodeTypeBizUpload] = newBizUploadNodeExecutor()
	engine.executors[NodeTypeBizMonitor] = newBizMonitorNodeExecutor()
	engine.executors[NodeTypeBizDeploy] = newBizDeployNodeExecutor()
	engine.executors[NodeTypeBizNotify] = newBizNotifyNodeExecutor()
	return engine
}
