package engine

import (
	"crypto/x509"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"slices"
	"strings"
	"time"

	legocertifier "github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	legolog "github.com/go-acme/lego/v4/log"
	"github.com/samber/lo"
	"github.com/xhit/go-str2duration/v2"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/certacme"
	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/repository"
	"github.com/certimate-go/certimate/internal/tools/mproc"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcryptokey "github.com/certimate-go/certimate/pkg/utils/crypto/key"
	xenv "github.com/certimate-go/certimate/pkg/utils/env"
)

var envMultiProc = true

func init() {
	envMultiProc = xenv.GetOrDefaultBool("CERTIMATE_WORKFLOW_MULTIPROC", true)
}

const (
	BizApplyKeySourceAuto   = "auto"
	BizApplyKeySourceReuse  = "reuse"
	BizApplyKeySourceCustom = "custom"
)

/**
 * Outputs:
 *   - ref: "certificate": string
 *
 * Variables:
 *   - "node.skipped": boolean
 *   - "certificate.commanName": string
 *   - "certificate.subjectAltNames": string
 *   - "certificate.notBefore": datetime
 *   - "certificate.notAfter": datetime
 *   - "certificate.hoursLeft": number
 *   - "certificate.daysLeft": number
 *   - "certificate.validity": boolean
 */
type bizApplyNodeExecutor struct {
	nodeExecutor

	accessRepo      accessRepository
	certificateRepo certificateRepository
	wfoutputRepo    workflowOutputRepository
}

func (ne *bizApplyNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	nodeCfg := execCtx.Node.Data.Config.AsBizApply()
	ne.logger.Info("ready to request certificate ...", slog.Any("config", nodeCfg))

	// 查询上次执行结果
	lastOutput, lastCertificate, err := ne.getLastOutputArtifacts(execCtx)
	if err != nil {
		return execRes, err
	} else {
		if lastOutput != nil {
			ne.logger.Info(fmt.Sprintf("found last run #%s", lastOutput.RunId))
		}

		if lastCertificate != nil {
			ne.setOuputsOfResult(execCtx, execRes, lastCertificate, false)
			ne.setVariablesOfResult(execCtx, execRes, lastCertificate)
		}
	}

	// 检测是否可以跳过本次执行
	if skippable, reason := ne.checkCanSkip(execCtx, lastOutput, lastCertificate); skippable {
		ne.logger.Info(fmt.Sprintf("skip this application, because %s", reason))

		execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyNodeSkipped, true, "boolean")
		return execRes, nil
	} else {
		if reason != "" {
			ne.logger.Info(fmt.Sprintf("re-apply, because %s", reason))
		} else {
			ne.logger.Info("no found last issued certificate, begin to apply")
		}

		execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyNodeSkipped, false, "boolean")
	}

	// 申请证书
	obtainResp, err := ne.executeObtain(execCtx, &nodeCfg, lastCertificate)
	if err != nil {
		return execRes, err
	}

	// 保存证书实体
	certificate := &domain.Certificate{
		Source:            domain.CertificateSourceTypeRequest,
		Certificate:       obtainResp.FullChainCertificate,
		PrivateKey:        obtainResp.PrivateKey,
		IssuerCertificate: obtainResp.IssuerCertificate,
		ACMEAcctUrl:       obtainResp.ACMEAcctUrl,
		ACMECertUrl:       obtainResp.ACMECertUrl,
		ACMECertStableUrl: obtainResp.ACMECertStableUrl,
		WorkflowId:        execCtx.WorkflowId,
		WorkflowRunId:     execCtx.RunId,
		WorkflowNodeId:    execCtx.Node.Id,
	}
	certificate.PopulateFromPEM(obtainResp.FullChainCertificate, obtainResp.PrivateKey)
	if certificate, err := ne.certificateRepo.Save(execCtx.ctx, certificate); err != nil {
		ne.logger.Warn("could not save certificate")
		return execRes, err
	} else {
		ne.logger.Info("certificate saved", slog.String("recordId", certificate.Id))
	}

	// 保存 ARI 替换状态
	if lastCertificate != nil && obtainResp.ARIReplaced {
		lastCertificate.IsRenewed = true
		ne.certificateRepo.Save(execCtx.ctx, lastCertificate)
	}

	// 节点输出
	ne.setOuputsOfResult(execCtx, execRes, certificate, true)
	ne.setVariablesOfResult(execCtx, execRes, certificate)

	ne.logger.Info("application completed")
	return execRes, nil
}

func (ne *bizApplyNodeExecutor) getLastOutputArtifacts(execCtx *NodeExecutionContext) (*domain.WorkflowOutput, *domain.Certificate, error) {
	lastOutput, err := ne.wfoutputRepo.GetByWorkflowIdAndNodeId(execCtx.ctx, execCtx.WorkflowId, execCtx.Node.Id)
	if err != nil && !domain.IsRecordNotFoundError(err) {
		return nil, nil, fmt.Errorf("failed to get last output record of node #%s: %w", execCtx.Node.Id, err)
	}

	if lastOutput != nil {
		lastCertificate, err := ne.certificateRepo.GetByWorkflowRunIdAndNodeId(execCtx.ctx, lastOutput.RunId, lastOutput.NodeId)
		if err != nil && !domain.IsRecordNotFoundError(err) {
			return lastOutput, nil, fmt.Errorf("failed to get last certificate record of node #%s: %w", execCtx.Node.Id, err)
		}

		return lastOutput, lastCertificate, nil
	}

	return lastOutput, nil, nil
}

func (ne *bizApplyNodeExecutor) checkCanSkip(execCtx *NodeExecutionContext, lastOutput *domain.WorkflowOutput, lastCertificate *domain.Certificate) (_skip bool, _reason string) {
	thisNodeCfg := execCtx.Node.Data.Config.AsBizApply()

	if lastOutput != nil && lastOutput.Succeeded {
		// 比较和上次申请时的关键配置（即影响证书签发的）参数是否一致
		lastNodeCfg := lastOutput.NodeConfig.AsBizApply()

		if !slices.Equal(thisNodeCfg.Domains, lastNodeCfg.Domains) {
			return false, "the configuration item 'Domains' changed"
		}
		if !slices.Equal(thisNodeCfg.IPAddrs, lastNodeCfg.IPAddrs) {
			return false, "the configuration item 'IPAddrs' changed"
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
		if thisNodeCfg.KeySource == BizApplyKeySourceCustom && thisNodeCfg.KeyContent != lastNodeCfg.KeyContent {
			return false, "the configuration item 'KeyContent' changed"
		}
		if thisNodeCfg.ValidityLifetime != lastNodeCfg.ValidityLifetime {
			return false, "the configuration item 'ValidityLifetime' changed"
		}
		if thisNodeCfg.PreferredChain != lastNodeCfg.PreferredChain {
			return false, "the configuration item 'PreferredChain' changed"
		}
		if thisNodeCfg.ACMEProfile != lastNodeCfg.ACMEProfile {
			return false, "the configuration item 'ACMEProfile' changed"
		}
		if thisNodeCfg.DisableCommonName != lastNodeCfg.DisableCommonName {
			return false, "the configuration item 'DisableCommonName' changed"
		}
	}

	if lastCertificate != nil {
		renewalInterval := time.Duration(thisNodeCfg.SkipBeforeExpiryDays) * time.Hour * 24
		expirationTime := time.Until(lastCertificate.ValidityNotAfter)
		daysLeft := int(math.Floor(expirationTime.Hours() / 24))
		if expirationTime > renewalInterval {
			return true, fmt.Sprintf("the last issued certificate #%s expires in %d day(s), next renewal will be in %d day(s)", lastCertificate.Id, daysLeft, thisNodeCfg.SkipBeforeExpiryDays)
		}

		return false, fmt.Sprintf("the last issued certificate #%s expires in %d day(s)", lastCertificate.Id, daysLeft)
	}

	return false, ""
}

func (ne *bizApplyNodeExecutor) executeObtain(execCtx *NodeExecutionContext, nodeCfg *domain.WorkflowNodeConfigForBizApply, lastCertificate *domain.Certificate) (*certacme.ObtainCertificateResponse, error) {
	// 读取私钥算法
	// 如果复用私钥，则保持算法一致
	legoKeyType, err := domain.CertificateKeyAlgorithmType(nodeCfg.KeyAlgorithm).KeyType()
	if err != nil {
		return nil, err
	} else {
		switch nodeCfg.KeySource {
		case BizApplyKeySourceAuto:
			break
		case BizApplyKeySourceReuse:
			if lastCertificate != nil {
				legoKeyType, _ = lastCertificate.KeyAlgorithm.KeyType()
			}
		case BizApplyKeySourceCustom:
			privkey, err := xcert.ParsePrivateKeyFromPEM(nodeCfg.KeyContent)
			if err != nil {
				return nil, fmt.Errorf("could not parse custom private key: %w", err)
			} else {
				privkeyAlg, privkeySize, _ := xcryptokey.GetPrivateKeyAlgorithm(privkey)
				switch privkeyAlg {
				case x509.RSA:
					if nodeCfg.KeyAlgorithm != fmt.Sprintf("RSA%d", privkeySize) {
						return nil, fmt.Errorf("could not parse custom private key: unsupported algorithm or key size")
					}
				case x509.ECDSA:
					if nodeCfg.KeyAlgorithm != fmt.Sprintf("EC%d", privkeySize) {
						return nil, fmt.Errorf("could not parse custom private key: unsupported algorithm or key size")
					}
				default:
					return nil, fmt.Errorf("could not parse custom private key: unsupported algorithm")
				}
			}
		}
	}

	// 读取质询提供商授权
	providerAccessConfig := make(map[string]any)
	if nodeCfg.ProviderAccessId != "" {
		if access, err := ne.accessRepo.GetById(execCtx.ctx, nodeCfg.ProviderAccessId); err != nil {
			return nil, fmt.Errorf("failed to get access #%s record: %w", nodeCfg.ProviderAccessId, err)
		} else {
			providerAccessConfig = access.Config
		}
	}

	// 读取证书颁发机构授权
	caAccessConfig := make(map[string]any)
	if nodeCfg.CAProviderAccessId != "" {
		if access, err := ne.accessRepo.GetById(execCtx.ctx, nodeCfg.CAProviderAccessId); err != nil {
			return nil, fmt.Errorf("failed to get access #%s record: %w", nodeCfg.CAProviderAccessId, err)
		} else {
			caAccessConfig = access.Config
		}
	}

	// 初始化 ACME 配置项
	legoOptions := &certacme.ACMEConfigOptions{
		CAProvider:       nodeCfg.CAProvider,
		CAAccessConfig:   caAccessConfig,
		CAProviderConfig: nodeCfg.CAProviderConfig,
		CertifierKeyType: legoKeyType,
	}
	legoConfig, err := certacme.NewACMEConfig(legoOptions)
	if err != nil {
		ne.logger.Warn("could not initialize acme config")
		return nil, err
	} else {
		ne.logger.Info("acme config initialized", slog.String("acmeDirUrl", legoConfig.CADirUrl))
	}

	// 初始化 ACME 账户
	// 注意此步骤仍需在主进程中进行，以保证并发安全
	legoUser, err := certacme.NewACMEAccountWithSingleFlight(legoConfig, nodeCfg.ContactEmail)
	if err != nil {
		ne.logger.Warn("could not initialize acme account")
		return nil, err
	} else {
		ne.logger.Info("acme account initialized", slog.String("acmeAcctUrl", legoUser.ACMEAcctUrl))
	}

	// 构造证书申请请求
	legoDomains := make([]string, 0)
	legoDomains = append(legoDomains, nodeCfg.Domains...)
	legoDomains = append(legoDomains, nodeCfg.IPAddrs...)
	obtainReq := &certacme.ObtainCertificateRequest{
		DomainOrIPs:    legoDomains,
		PrivateKeyType: legoKeyType,
		PrivateKeyPEM: lo.
			If(nodeCfg.KeySource == BizApplyKeySourceAuto, "").
			ElseF(func() string {
				switch nodeCfg.KeySource {
				case BizApplyKeySourceReuse:
					if lastCertificate != nil {
						return lastCertificate.PrivateKey
					}
				case BizApplyKeySourceCustom:
					return nodeCfg.KeyContent
				}
				return ""
			}),
		ValidityNotAfter: lo.
			If(nodeCfg.ValidityLifetime == "", time.Time{}).
			ElseF(func() time.Time {
				duration, err := str2duration.ParseDuration(nodeCfg.ValidityLifetime)
				if err != nil {
					return time.Time{}
				}
				return time.Now().Add(duration)
			}),
		NoCommonName:           nodeCfg.DisableCommonName,
		ChallengeType:          nodeCfg.ChallengeType,
		Provider:               nodeCfg.Provider,
		ProviderAccessConfig:   providerAccessConfig,
		ProviderExtendedConfig: nodeCfg.ProviderConfig,
		DisableFollowCNAME:     nodeCfg.DisableFollowCNAME,
		Nameservers:            nodeCfg.Nameservers,
		DnsPropagationWait:     nodeCfg.DnsPropagationWait,
		DnsPropagationTimeout:  nodeCfg.DnsPropagationTimeout,
		DnsTTL:                 nodeCfg.DnsTTL,
		HttpDelayWait:          nodeCfg.HttpDelayWait,
		PreferredChain:         nodeCfg.PreferredChain,
		ACMEProfile:            nodeCfg.ACMEProfile,
		ARIReplacesAcctUrl: lo.
			If(nodeCfg.DisableARI || lastCertificate == nil, "").
			ElseF(func() string {
				if lastCertificate.IsRenewed {
					return ""
				}
				return lastCertificate.ACMEAcctUrl
			}),
		ARIReplacesCertId: lo.
			If(nodeCfg.DisableARI || lastCertificate == nil, "").
			ElseF(func() string {
				if lastCertificate.IsRenewed {
					return ""
				}

				newCertSan := slices.Clone(nodeCfg.Domains)
				oldCertSan := strings.Split(lastCertificate.SubjectAltNames, ";")
				slices.Sort(newCertSan)
				slices.Sort(oldCertSan)
				if !slices.Equal(newCertSan, oldCertSan) {
					return ""
				}

				oldCertX509, err := xcert.ParseCertificateFromPEM(lastCertificate.Certificate)
				if err != nil {
					return ""
				}

				oldARICertId, _ := legocertifier.MakeARICertID(oldCertX509)
				return oldARICertId
			}),
	}

	// 如果启用多进程模式，发送指令
	if envMultiProc {
		type InData struct {
			Account *certacme.ACMEAccount              `json:"account,omitempty"`
			Request *certacme.ObtainCertificateRequest `json:"request,omitempty"`
		}

		type OutData struct {
			Response *certacme.ObtainCertificateResponse `json:"response"`
		}

		msender := mproc.NewSender[InData, OutData]("certapply", ne.logger)
		moutput, err := msender.SendWithContext(execCtx.ctx, &InData{
			Account: legoUser,
			Request: obtainReq,
		})
		if err != nil {
			ne.logger.Warn("could not obtain certificate")
			return nil, err
		}

		if moutput.Response != nil {
			return moutput.Response, nil
		} else {
			panic("unreachable")
		}
	}

	// 初始化 ACME 客户端
	legolog.Logger = certacme.NewLegoLogger(app.GetLogger())
	legoClient, err := certacme.NewACMEClientWithAccount(legoUser, func(c *lego.Config) error {
		c.UserAgent = "certimate"
		c.Certificate.KeyType = legoKeyType
		c.Certificate.DisableCommonName = obtainReq.NoCommonName
		return nil
	})
	if err != nil {
		ne.logger.Warn("could not initialize acme client")
		return nil, err
	}

	// 执行申请证书请求
	obtainResp, err := legoClient.ObtainCertificate(execCtx.ctx, obtainReq)
	if err != nil {
		ne.logger.Warn("could not obtain certificate")
		return nil, err
	}

	return obtainResp, nil
}

func (ne *bizApplyNodeExecutor) setOuputsOfResult(execCtx *NodeExecutionContext, execRes *NodeExecutionResult, certificate *domain.Certificate, persistent bool) {
	if certificate != nil {
		key := "certificate"
		value := fmt.Sprintf("%s#%s", domain.CollectionNameCertificate, certificate.Id)
		if persistent {
			execRes.AddOutputWithPersistent(stateIOTypeRef, key, value, "string")
		} else {
			execRes.AddOutput(stateIOTypeRef, key, value, "string")
		}
	}
}

func (ne *bizApplyNodeExecutor) setVariablesOfResult(execCtx *NodeExecutionContext, execRes *NodeExecutionResult, certificate *domain.Certificate) {
	var vCommonName string
	var vSubjectAltNames string
	var vNotBefore time.Time
	var vNotAfter time.Time
	var vHoursLeft int32
	var vDaysLeft int32
	var vValidity bool

	if certificate != nil {
		vCommonName = strings.Split(certificate.SubjectAltNames, ";")[0]
		vSubjectAltNames = certificate.SubjectAltNames
		vNotBefore = certificate.ValidityNotBefore
		vNotAfter = certificate.ValidityNotAfter
		vHoursLeft = int32(math.Floor(time.Until(certificate.ValidityNotAfter).Hours()))
		vDaysLeft = int32(math.Floor(time.Until(certificate.ValidityNotAfter).Hours() / 24))
		vValidity = certificate.ValidityNotAfter.After(time.Now())
	}

	execRes.AddVariable(stateVarKeyCertificateDomain, vCommonName, "string")
	execRes.AddVariable(stateVarKeyCertificateDomains, vSubjectAltNames, "string")
	execRes.AddVariable(stateVarKeyCertificateCommonName, vCommonName, "string")
	execRes.AddVariable(stateVarKeyCertificateSubjectAltNames, vSubjectAltNames, "string")
	execRes.AddVariable(stateVarKeyCertificateNotBefore, vNotBefore, "datetime")
	execRes.AddVariable(stateVarKeyCertificateNotAfter, vNotAfter, "datetime")
	execRes.AddVariable(stateVarKeyCertificateHoursLeft, vHoursLeft, "number")
	execRes.AddVariable(stateVarKeyCertificateDaysLeft, vDaysLeft, "number")
	execRes.AddVariable(stateVarKeyCertificateValidity, vValidity, "boolean")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateDomain, vCommonName, "string")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateDomains, vSubjectAltNames, "string")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateCommonName, vCommonName, "string")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateSubjectAltNames, vSubjectAltNames, "string")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateNotBefore, vNotBefore, "datetime")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateNotAfter, vNotAfter, "datetime")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateHoursLeft, vHoursLeft, "number")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateDaysLeft, vDaysLeft, "number")
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyCertificateValidity, vValidity, "boolean")
}

func newBizApplyNodeExecutor() NodeExecutor {
	return &bizApplyNodeExecutor{
		nodeExecutor:    nodeExecutor{logger: slog.Default()},
		accessRepo:      repository.NewAccessRepository(),
		certificateRepo: repository.NewCertificateRepository(),
		wfoutputRepo:    repository.NewWorkflowOutputRepository(),
	}
}
