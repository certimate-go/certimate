package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/certapply"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/settings"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

const tencentCloudPendingCertificateKey = "tencentcloudPendingCertificateId"

type tencentCloudIssuer interface {
	Obtain(context.Context, certapply.TencentCloudRequest, func(string) error) (*certapply.TencentCloudResult, error)
}

type tencentCloudClientFactory func(certapply.TencentCloudConfig, *slog.Logger) (tencentCloudIssuer, error)

func newTencentCloudIssuer(config certapply.TencentCloudConfig, logger *slog.Logger) (tencentCloudIssuer, error) {
	return certapply.NewTencentCloudClient(config, logger)
}

func tencentCloudRequestFromNode(nodeCfg *domain.WorkflowNodeConfigForBizApply) certapply.TencentCloudRequest {
	method := nodeCfg.ChallengeType
	switch method {
	case "", "dns-auto":
		method = "DNS_AUTO"
	case "dns-01":
		method = "DNS"
	}
	return certapply.TencentCloudRequest{
		Domains:            nodeCfg.Domains,
		IPAddrs:            nodeCfg.IPAddrs,
		ContactEmail:       nodeCfg.ContactEmail,
		KeySource:          nodeCfg.KeySource,
		KeyAlgorithm:       nodeCfg.KeyAlgorithm,
		DvAuthMethod:       method,
		DNSPropagationWait: nodeCfg.DnsPropagationWait,
	}
}

func tencentCloudPendingCertificate(nodeCfg *domain.WorkflowNodeConfigForBizApply, lastOutput *domain.WorkflowOutput) string {
	if lastOutput == nil || lastOutput.Succeeded {
		return ""
	}
	lastCfg := lastOutput.NodeConfig.AsBizApply()
	if lastCfg.Source != "tencentcloud" || lastCfg.TencentCloudAccessId != nodeCfg.TencentCloudAccessId ||
		!slices.Equal(lastCfg.Domains, nodeCfg.Domains) || !slices.Equal(lastCfg.IPAddrs, nodeCfg.IPAddrs) ||
		lastCfg.KeyAlgorithm != nodeCfg.KeyAlgorithm || lastCfg.ContactEmail != nodeCfg.ContactEmail ||
		tencentCloudRequestFromNode(&lastCfg).DvAuthMethod != tencentCloudRequestFromNode(nodeCfg).DvAuthMethod {
		return ""
	}
	for _, output := range lastOutput.Outputs {
		if output != nil && output.Name == tencentCloudPendingCertificateKey {
			return output.Value
		}
	}
	return ""
}

func (ne *bizApplyNodeExecutor) execObtainTencentCloudCertificate(execCtx *NodeExecutionContext, execRes *NodeExecutionResult, nodeCfg *domain.WorkflowNodeConfigForBizApply, lastOutput *domain.WorkflowOutput) (*certacme.ObtainCertificateResponse, error) {
	access, err := ne.accessRepo.GetById(execCtx.Context(), nodeCfg.TencentCloudAccessId)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tencent Cloud access #%s: %w", nodeCfg.TencentCloudAccessId, err)
	}
	if access.Provider != domain.AccessProviderTypeTencentCloud.String() {
		return nil, fmt.Errorf("access #%s is not a Tencent Cloud credential", nodeCfg.TencentCloudAccessId)
	}
	credentials := domain.AccessConfigForTencentCloud{}
	if err := xmaps.Populate(access.Config, &credentials); err != nil {
		return nil, fmt.Errorf("invalid Tencent Cloud credentials: %w", err)
	}
	var dnsProvider certapply.DNSProvider
	if nodeCfg.ChallengeType == "dns-01" {
		dnsAccess, err := ne.accessRepo.GetById(execCtx.Context(), nodeCfg.ProviderAccessId)
		if err != nil {
			return nil, fmt.Errorf("failed to get DNS access #%s: %w", nodeCfg.ProviderAccessId, err)
		}
		provider := strings.TrimSuffix(nodeCfg.Provider, "-dns")
		if dnsAccess.Provider != provider || dnsAccess.Reserve != "" {
			return nil, fmt.Errorf("DNS access #%s does not match provider %q", nodeCfg.ProviderAccessId, nodeCfg.Provider)
		}
		dnsProvider, err = certapply.NewDNSProvider(certapply.DNSProviderConfig{
			Provider: provider, AccessConfig: dnsAccess.Config, TTL: nodeCfg.DnsTTL,
		})
		if err != nil {
			return nil, err
		}
	}
	client, err := ne.newTencentCloudClient(certapply.TencentCloudConfig{
		SecretId:    credentials.SecretId,
		SecretKey:   credentials.SecretKey,
		ProjectId:   credentials.ProjectId,
		DNSProvider: dnsProvider,
	}, ne.logger)
	if err != nil {
		return nil, err
	}

	timeout := 10 * time.Minute
	if seconds := settings.GetGlobalSettingsForSSLProvider().Timeout; seconds > 0 {
		timeout = time.Duration(seconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(execCtx.Context(), timeout)
	defer cancel()

	request := tencentCloudRequestFromNode(nodeCfg)
	request.CertificateId = tencentCloudPendingCertificate(nodeCfg, lastOutput)
	pendingOutput := &domain.WorkflowOutput{
		WorkflowId: execCtx.WorkflowId,
		RunId:      execCtx.RunId,
		NodeId:     execCtx.Node.Id,
		NodeConfig: execCtx.Node.Data.Config,
		Succeeded:  false,
	}
	persistOrder := func(id string) error {
		pendingOutput.Outputs = []*domain.WorkflowOutputEntry{{
			Type:      stateIOTypePlain,
			Name:      tencentCloudPendingCertificateKey,
			Value:     id,
			ValueType: stateValTypeString,
		}}
		_, err := ne.wfoutputRepo.Save(execCtx.Context(), pendingOutput)
		return err
	}
	if request.CertificateId != "" {
		if err := persistOrder(request.CertificateId); err != nil {
			return nil, fmt.Errorf("failed to persist resumed Tencent Cloud order: %w", err)
		}
	}
	result, err := client.Obtain(ctx, request, persistOrder)
	if err != nil {
		var statusErr *certapply.TencentCloudStatusError
		if errors.As(err, &statusErr) && statusErr.IsTerminal() {
			pendingOutput.Outputs = nil
			if _, saveErr := ne.wfoutputRepo.Save(execCtx.Context(), pendingOutput); saveErr != nil {
				return nil, errors.Join(err, fmt.Errorf("failed to clear rejected Tencent Cloud order: %w", saveErr))
			}
		}
		return nil, err
	}

	execRes.AddOutputWithPersistent(stateIOTypePlain, "tencentcloudCertificateId", result.CertificateId, stateValTypeString)
	return &certacme.ObtainCertificateResponse{
		CAProvider:           domain.CAProviderType("tencentcloud"),
		FullChainCertificate: result.Certificate,
		PrivateKey:           result.PrivateKey,
	}, nil
}
