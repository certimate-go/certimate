package zenlayerga

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	zcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/zenlayer-ga"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/zenlayer-ga/internal"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// Zenlayer AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// Zenlayer AccessKeyPassword。
	AccessKeyPassword string `json:"accessKeyPassword"`
	// Zenlayer 资源组 ID。
	ResourceGroupId string `json:"resourceGroupId,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 加速器 ID。
	// 部署资源类型为 [RESOURCE_TYPE_ACCELERATOR] 时必填。
	AcceleratorId string `json:"acceleratorId,omitempty"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *internal.ZgaClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.AccessKeyPassword)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_ACCELERATOR:
		if err := d.deployToAccelerator(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToAccelerator(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.AcceleratorId == "" {
		return errors.New("config `acceleratorId` is required")
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
	describeAcceleratorsReq := zga.NewDescribeAcceleratorsRequest()
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
		modifyAcceleratorCertificateReq := zga.NewModifyAcceleratorCertificateRequest()
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
		describeAcceleratorsReq := zga.NewDescribeAcceleratorsRequest()
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
	}, time.Second*5); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
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

func createSDKClient(accessKeyId, accessKeyPassword string) (*internal.ZgaClient, error) {
	config := zcommon.NewConfig()

	client, err := internal.NewZgaClient(config, accessKeyId, accessKeyPassword)
	if err != nil {
		return nil, err
	}

	return client, nil
}
