package nodeprocessor

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"golang.org/x/exp/maps"

	"github.com/certimate-go/certimate/internal/applicant"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type applyNode struct {
	node *domain.WorkflowNode
	*nodeProcessor
	*nodeOutputer

	certRepo   certificateRepository
	outputRepo workflowOutputRepository
}

func NewApplyNode(node *domain.WorkflowNode) *applyNode {
	return &applyNode{
		node:          node,
		nodeProcessor: newNodeProcessor(node),
		nodeOutputer:  newNodeOutputer(),

		certRepo:   repository.NewCertificateRepository(),
		outputRepo: repository.NewWorkflowOutputRepository(),
	}
}

func (n *applyNode) Process(ctx context.Context) error {
	nodeCfg := n.node.GetConfigForApply()
	n.logger.Info("ready to obtain certificate ...", slog.Any("config", nodeCfg))

	// 查询上次执行结果
	lastOutput, err := n.outputRepo.GetByNodeId(ctx, n.node.Id)
	if err != nil && !domain.IsRecordNotFoundError(err) {
		return err
	}

	// 检测是否可以跳过本次执行
	if skippable, reason := n.checkCanSkip(ctx, lastOutput); skippable {
		n.outputs[outputKeyForNodeSkipped] = strconv.FormatBool(true)
		n.logger.Info(fmt.Sprintf("skip this application, because %s", reason))
		return nil
	} else if reason != "" {
		n.logger.Info(fmt.Sprintf("re-apply, because %s", reason))
	}

	// 初始化申请器
	applicant, err := applicant.NewWithWorkflowNode(applicant.ApplicantWithWorkflowNodeConfig{
		Node:   n.node,
		Logger: n.logger,
	})
	if err != nil {
		n.logger.Warn("failed to create applicant provider")
		return err
	}

	// 申请证书
	applyResult, err := applicant.Apply(ctx)
	if err != nil {
		n.logger.Warn("failed to obtain certificate")
		return err
	}

	// 解析证书并生成实体
	certX509, err := xcert.ParseCertificateFromPEM(applyResult.FullChainCertificate)
	if err != nil {
		n.logger.Warn("failed to parse certificate, may be the CA responded error")
		return err
	}

	certificate := &domain.Certificate{
		Source:            domain.CertificateSourceTypeWorkflow,
		Certificate:       applyResult.FullChainCertificate,
		PrivateKey:        applyResult.PrivateKey,
		IssuerCertificate: applyResult.IssuerCertificate,
		ACMEAccountUrl:    applyResult.ACMEAccountUrl,
		ACMECertUrl:       applyResult.ACMECertUrl,
		ACMECertStableUrl: applyResult.ACMECertStableUrl,
	}
	certificate.PopulateFromX509(certX509)

	// 保存执行结果
	output := &domain.WorkflowOutput{
		WorkflowId: getContextWorkflowId(ctx),
		RunId:      getContextWorkflowRunId(ctx),
		NodeId:     n.node.Id,
		Node:       n.node,
		Succeeded:  true,
		Outputs:    n.node.Outputs,
	}
	if _, err := n.outputRepo.SaveWithCertificate(ctx, output, certificate); err != nil {
		n.logger.Warn("failed to save node output")
		return err
	}

	// 保存 ARI 记录
	if applyResult.ARIReplaced && lastOutput != nil {
		lastCertificate, _ := n.certRepo.GetByWorkflowRunIdAndNodeId(ctx, lastOutput.RunId, lastOutput.NodeId)
		if lastCertificate != nil {
			lastCertificate.ACMERenewed = true
			n.certRepo.Save(ctx, lastCertificate)
		}
	}

	// 记录中间结果
	n.outputs[outputKeyForNodeSkipped] = strconv.FormatBool(false)
	n.outputs[outputKeyForCertificateValidity] = strconv.FormatBool(true)
	n.outputs[outputKeyForCertificateDaysLeft] = strconv.FormatInt(int64(time.Until(certificate.ExpireAt).Hours()/24), 10)

	n.logger.Info("application completed")
	return nil
}

func (n *applyNode) checkCanSkip(ctx context.Context, lastOutput *domain.WorkflowOutput) (_skip bool, _reason string) {
	if lastOutput != nil && lastOutput.Succeeded {
		// 比较和上次申请时的关键配置（即影响证书签发的）参数是否一致
		thisNodeCfg := n.node.GetConfigForApply()
		lastNodeCfg := lastOutput.Node.GetConfigForApply()

		if thisNodeCfg.Domains != lastNodeCfg.Domains {
			return false, "the configuration item 'Domains' changed"
		}
		if thisNodeCfg.ContactEmail != lastNodeCfg.ContactEmail {
			return false, "the configuration item 'ContactEmail' changed"
		}
		if thisNodeCfg.Provider != lastNodeCfg.Provider {
			return false, "the configuration item 'Provider' changed"
		}
		if thisNodeCfg.ProviderAccessId != lastNodeCfg.ProviderAccessId {
			return false, "the configuration item 'ProviderAccessId' changed"
		}
		if !maps.Equal(thisNodeCfg.ProviderConfig, lastNodeCfg.ProviderConfig) {
			return false, "the configuration item 'ProviderConfig' changed"
		}
		if thisNodeCfg.CAProvider != lastNodeCfg.CAProvider {
			return false, "the configuration item 'CAProvider' changed"
		}
		if thisNodeCfg.CAProviderAccessId != lastNodeCfg.CAProviderAccessId {
			return false, "the configuration item 'CAProviderAccessId' changed"
		}
		if !maps.Equal(thisNodeCfg.CAProviderConfig, lastNodeCfg.CAProviderConfig) {
			return false, "the configuration item 'CAProviderConfig' changed"
		}
		if thisNodeCfg.KeyAlgorithm != lastNodeCfg.KeyAlgorithm {
			return false, "the configuration item 'KeyAlgorithm' changed"
		}

		lastCertificate, _ := n.certRepo.GetByWorkflowRunIdAndNodeId(ctx, lastOutput.RunId, lastOutput.NodeId)
		if lastCertificate != nil {
			renewalInterval := time.Duration(thisNodeCfg.SkipBeforeExpiryDays) * time.Hour * 24
			expirationTime := time.Until(lastCertificate.ExpireAt)
			if expirationTime > renewalInterval {
				daysLeft := int(expirationTime.Hours() / 24)
				// TODO: 优化此处逻辑，[checkCanSkip] 方法不应该修改中间结果，违背单一职责
				n.outputs[outputKeyForCertificateValidity] = strconv.FormatBool(true)
				n.outputs[outputKeyForCertificateDaysLeft] = strconv.FormatInt(int64(daysLeft), 10)

				return true, fmt.Sprintf("the certificate has already been issued (expires in %d day(s), next renewal in %d day(s))", daysLeft, thisNodeCfg.SkipBeforeExpiryDays)
			}
		}
	}

	return false, ""
}
