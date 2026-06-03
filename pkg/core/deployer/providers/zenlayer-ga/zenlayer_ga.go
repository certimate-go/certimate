package zenlayerga

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	zcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"

	"github.com/certimate-go/certimate/pkg/core"
	cmgrimpl "github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-ga"
	zgasdk "github.com/certimate-go/certimate/pkg/sdk3rd/zenlayer/zga"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// Zenlayer AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// Zenlayer AccessKeyPassword。
	AccessKeyPassword string `json:"accessKeyPassword"`
	// Zenlayer 资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 加速器 ID。
	// 部署目标为 [DEPLOY_TARGET_ACCELERATOR] 时必填。
	AcceleratorId string `json:"acceleratorId,omitempty"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *zgasdk.Client
	sdkCertmgr core.Certmgr
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := cmgrimpl.NewCertmgr(&cmgrimpl.CertmgrConfig{
		AccessKeyId:       config.AccessKeyId,
		AccessKeyPassword: config.AccessKeyPassword,
		ResourceGroupId:   config.ResourceGroupId,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create certmgr: %w", err)
	}

	return &Deployer{
		config:     config,
		logger:     slog.Default(),
		sdkClient:  client,
		sdkCertmgr: pcertmgr,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}

	d.sdkCertmgr.SetLogger(logger)
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_ACCELERATOR:
		if err := d.deployToAccelerator(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToAccelerator(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.AcceleratorId == "" {
		return fmt.Errorf("config `acceleratorId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 查询加速器信息
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/zga/accelerator/describeaccelerators
	describeAcceleratorsReq := zgasdk.NewDescribeAcceleratorsRequest()
	describeAcceleratorsReq.AcceleratorIds = []string{d.config.AcceleratorId}
	describeAcceleratorsReq.PageNum = 1
	describeAcceleratorsReq.PageSize = 1
	describeAcceleratorsResp, err := d.sdkClient.DescribeAccelerators(describeAcceleratorsReq)
	d.logger.Debug("sdk request 'zga.DescribeAccelerators'", slog.Any("request", describeAcceleratorsReq), slog.Any("response", describeAcceleratorsResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'zga.DescribeAccelerators': %w", err)
	} else if len(describeAcceleratorsResp.Response.DataSet) == 0 {
		return fmt.Errorf("could not found accelerator '%s'", d.config.AcceleratorId)
	}

	// 修改加速器证书
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/zga/accelerator/modifyacceleratorcertificate
	acceleratorInfo := describeAcceleratorsResp.Response.DataSet[0]
	if acceleratorInfo.Certificate == nil || acceleratorInfo.Certificate.CertificateId != upres.CertId {
		modifyAcceleratorCertificateReq := zgasdk.NewModifyAcceleratorCertificateRequest()
		modifyAcceleratorCertificateReq.AcceleratorId = acceleratorInfo.AcceleratorId
		modifyAcceleratorCertificateReq.CertificateId = upres.CertId
		modifyAcceleratorCertificateResp, err := d.sdkClient.ModifyAcceleratorCertificate(modifyAcceleratorCertificateReq)
		d.logger.Debug("sdk request 'zga.ModifyAcceleratorCertificate'", slog.Any("request", modifyAcceleratorCertificateReq), slog.Any("response", modifyAcceleratorCertificateResp))
		if err != nil {
			return fmt.Errorf("failed to execute sdk request 'zga.ModifyAcceleratorCertificate': %w", err)
		}
	}

	// 查询加速器状态，等待部署状态变更
	// REF: https://docs.console.zenlayer.com/api-reference/cn/networking/zga/accelerator/describeaccelerators
	if _, err := xwait.UntilWithContext(ctx, func(_ context.Context, _ int) (bool, error) {
		describeAcceleratorsReq := zgasdk.NewDescribeAcceleratorsRequest()
		describeAcceleratorsReq.AcceleratorIds = []string{acceleratorInfo.AcceleratorId}
		describeAcceleratorsReq.PageNum = 1
		describeAcceleratorsReq.PageSize = 1
		describeAcceleratorsResp, err := d.sdkClient.DescribeAccelerators(describeAcceleratorsReq)
		d.logger.Debug("sdk request 'zga.DescribeAccelerators'", slog.Any("request", describeAcceleratorsReq), slog.Any("response", describeAcceleratorsResp))
		if err != nil {
			return false, fmt.Errorf("failed to execute sdk request 'zga.DescribeAccelerators': %w", err)
		} else if len(describeAcceleratorsResp.Response.DataSet) == 0 {
			return false, fmt.Errorf("could not found accelerator '%s'", d.config.AcceleratorId)
		}

		switch describeAcceleratorsResp.Response.DataSet[0].AcceleratorStatus {
		case "Accelerating":
			return true, nil
		case "NotAccelerate", "StopAccelerate", "AccelerateFailure":
			return false, fmt.Errorf("unexpected zenlayer accelerator status")
		}

		d.logger.Info("waiting for zenlayer accelerator deploying completion ...")
		return false, nil
	}, 10*time.Second); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 替换证书
	upres, err := d.sdkCertmgr.Replace(ctx, d.config.CertificateId, certPEM, privkeyPEM)
	if err != nil {
		return fmt.Errorf("failed to replace certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate replaced", slog.Any("result", upres))
	}

	return nil
}

func createSDKClient(accessKeyId, accessKeyPassword string) (*zgasdk.Client, error) {
	config := zcommon.NewConfig()

	client, err := zgasdk.NewClient(config, accessKeyId, accessKeyPassword)
	if err != nil {
		return nil, err
	}

	return client, nil
}
