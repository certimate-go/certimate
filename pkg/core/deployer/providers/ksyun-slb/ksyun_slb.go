package ksyunslb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/KscSDK/ksc-sdk-go/ksc"
	ksckcm "github.com/KscSDK/ksc-sdk-go/service/kcm"

	"github.com/certimate-go/certimate/pkg/core/deployer"
)

type DeployerConfig struct {
	// 金山云 AccessKeyId。
	AccessKeyId string `json:"accessKeyId"`
	// 金山云 SecretAccessKey。
	SecretAccessKey string `json:"secretAccessKey"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId string `json:"certificateId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *ksckcm.Kcm
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.AccessKeyId, config.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	return &Deployer{
		config:    config,
		logger:    slog.Default(),
		sdkClient: client,
	}, nil
}

func (d *Deployer) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 根据部署资源类型决定部署方式
	switch d.config.ResourceType {
	case RESOURCE_TYPE_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported resource type '%s'", d.config.ResourceType)
	}

	return &deployer.DeployResult{}, nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == "" {
		return errors.New("config `certificateId` is required")
	}

	// 更新证书信息
	// https://docs.ksyun.com/documents/2121
	modifyCertificateInput := map[string]any{
		"CertificateId":   d.config.CertificateId,
		"CertificateName": fmt.Sprintf("certimate_%d", time.Now().UnixMilli()),
		"PublicKey":       certPEM,
		"PrivateKey":      privkeyPEM,
	}
	modifyCertificateOutput, err := d.sdkClient.ModifyCertificateWithContext(ctx, &modifyCertificateInput)
	d.logger.Debug("sdk request 'kcm.ModifyCertificate'", slog.Any("request", modifyCertificateInput), slog.Any("response", modifyCertificateOutput))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'kcm.ModifyCertificate': %w", err)
	}

	return nil
}

func createSDKClient(accessKeyId, secretAccessKey string) (*ksckcm.Kcm, error) {
	region := "cn-beijing-6"
	client := ksckcm.SdkNew(ksc.NewClient(accessKeyId, secretAccessKey), &ksc.Config{Region: &region})
	return client, nil
}
