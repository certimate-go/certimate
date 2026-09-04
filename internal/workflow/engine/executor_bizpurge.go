package engine

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/certimate-go/certimate/internal/certmgmt"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
)

/**
 * Variables:
 *   - "node.skipped": boolean
 */
type bizPurgeNodeExecutor struct {
	nodeExecutor

	accessRepo accessRepository
}

func (ne *bizPurgeNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	nodeCfg := execCtx.Node.Data.Config.AsBizPurge()
	ne.logger.Info("ready to purge certificate ...", slog.Any("config", nodeCfg))

	// 读取清除提供商授权
	providerAccessConfig := make(map[string]any)
	if nodeCfg.ProviderAccessId != "" {
		if access, err := ne.accessRepo.GetById(execCtx.Context(), nodeCfg.ProviderAccessId); err != nil {
			return nil, fmt.Errorf("failed to get access #%s record: %w", nodeCfg.ProviderAccessId, err)
		} else {
			providerAccessConfig = access.Config
		}
	}

	// 清除过期证书
	purger := certmgmt.NewClient(certmgmt.WithLogger(ne.logger))
	purgeReq := &certmgmt.PurgeCertificateRequest{
		Provider:               domain.PurgeProviderType(nodeCfg.Provider),
		ProviderAccessConfig:   providerAccessConfig,
		ProviderExtendedConfig: nodeCfg.ProviderConfig,
		Expiry:                 time.Duration(nodeCfg.ExpiredDays) * time.Hour * 24,
	}
	if _, err := purger.PurgeCertificate(execCtx.Context(), purgeReq); err != nil {
		ne.logger.Warn("could not purge certificate")
		return execRes, err
	}

	// 节点输出
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyNodeSkipped, false, stateValTypeBoolean)

	ne.logger.Info("purge completed")
	return execRes, nil
}

func newBizPurgeNodeExecutor() NodeExecutor {
	return &bizPurgeNodeExecutor{
		nodeExecutor: nodeExecutor{logger: slog.Default()},
		accessRepo:   repository.NewAccessRepository(),
	}
}
