package engine

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/notify"
	"github.com/certimate-go/certimate/internal/repository"
)

type bizNotifyNodeExecutor struct {
	nodeExecutor

	settingsRepo settingsRepository
	accessRepo   accessRepository
}

func (ne *bizNotifyNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	nodeCfg := execCtx.Node.Data.Config.AsBizNotify()
	ne.logger.Info("ready to send notification ...", slog.Any("config", nodeCfg))

	// 检测是否可以跳过本次执行
	if skippable, reason := ne.checkCanSkip(execCtx); skippable {
		ne.logger.Info(fmt.Sprintf("skip this application, because %s", reason))
		return execRes, nil
	}

	// 读取部署提供商授权
	providerAccessConfig := make(map[string]any)
	if nodeCfg.ProviderAccessId != "" {
		if access, err := ne.accessRepo.GetById(execCtx.Context(), nodeCfg.ProviderAccessId); err != nil {
			return nil, fmt.Errorf("failed to get access #%s record: %w", nodeCfg.ProviderAccessId, err)
		} else {
			providerAccessConfig = access.Config
		}
	}

	// 渲染通知模板
	reMustache := regexp.MustCompile(`\{\{\s*(\$[^\s]+)\s*\}\}`)
	reMustacheReplacer := func(match string) string {
		mustache := strings.TrimSpace(match[2 : len(match)-2])
		if mustache == "" {
			return match
		}

		key := mustache[1:]
		if key == "" {
			return match
		} else if key == "now" {
			return time.Now().Format(time.RFC3339)
		}

		// TODO: 支持作用域变量
		if state, ok := execCtx.variables.Get(key); ok {
			return state.ValueString()
		}

		return match
	}
	subject := reMustache.ReplaceAllStringFunc(nodeCfg.Subject, reMustacheReplacer)
	message := reMustache.ReplaceAllStringFunc(nodeCfg.Message, reMustacheReplacer)

	// 推送通知
	notifier := notify.NewClient(notify.WithLogger(ne.logger))
	notifyReq := &notify.SendNotificationRequest{
		Provider:               nodeCfg.Provider,
		ProviderAccessConfig:   providerAccessConfig,
		ProviderExtendedConfig: nodeCfg.ProviderConfig,
		Subject:                subject,
		Message:                message,
	}
	if _, err := notifier.SendNotification(execCtx.Context(), notifyReq); err != nil {
		ne.logger.Warn("could not send notification")
		return execRes, err
	}

	ne.logger.Info("notification completed")
	return execRes, nil
}

func (ne *bizNotifyNodeExecutor) checkCanSkip(execCtx *NodeExecutionContext) (_skip bool, _reason string) {
	thisNodeCfg := execCtx.Node.Data.Config.AsBizNotify()
	if !thisNodeCfg.SkipOnAllPrevSkipped {
		return false, ""
	}

	var total, skipped int32
	for _, variable := range execCtx.variables.All() {
		if variable.Scope != "" && variable.Key == stateVarKeyNodeSkipped {
			total++
			if variable.Value == true {
				skipped++
			}
		}
	}
	if total == 0 || skipped != total {
		return false, ""
	}

	return true, "all the previous nodes have been skipped"
}

func newBizNotifyNodeExecutor() NodeExecutor {
	return &bizNotifyNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
		settingsRepo: repository.NewSettingsRepository(),
		accessRepo:   repository.NewAccessRepository(),
	}
}
