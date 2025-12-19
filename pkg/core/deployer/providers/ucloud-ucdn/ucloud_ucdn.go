package uclouducdn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-ussl"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/ucdn"
)

type DeployerConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 加速域名 ID。
	DomainId string `json:"domainId"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ucloudsdk.UCDNClient
	sdkCertmgr certmgr.Provider
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.PrivateKey, config.PublicKey, config.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	pcertmgr, err := mcertmgr.NewCertmgr(&mcertmgr.CertmgrConfig{
		PrivateKey: config.PrivateKey,
		PublicKey:  config.PublicKey,
		ProjectId:  config.ProjectId,
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
	if d.config.DomainId == "" {
		return nil, errors.New("config `domainId` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 获取加速域名配置
	// REF: https://docs.ucloud.cn/api/ucdn-api/get_ucdn_domain_config
	getUcdnDomainConfigReq := d.sdkClient.NewGetUcdnDomainConfigRequest()
	getUcdnDomainConfigReq.DomainId = []string{d.config.DomainId}
	getUcdnDomainConfigResp, err := d.sdkClient.GetUcdnDomainConfig(getUcdnDomainConfigReq)
	d.logger.Debug("sdk request 'ucdn.GetUcdnDomainConfig'", slog.Any("request", getUcdnDomainConfigReq), slog.Any("response", getUcdnDomainConfigResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ucdn.GetUcdnDomainConfig': %w", err)
	} else if len(getUcdnDomainConfigResp.DomainList) == 0 {
		return nil, fmt.Errorf("could not find domain '%s'", d.config.DomainId)
	}

	// 更新 HTTPS 加速配置
	// REF: https://docs.ucloud.cn/api/ucdn-api/update_ucdn_domain_https_config_v2
	certId, _ := strconv.Atoi(upres.CertId)
	updateUcdnDomainHttpsConfigV2Req := d.sdkClient.NewUpdateUcdnDomainHttpsConfigV2Request()
	updateUcdnDomainHttpsConfigV2Req.DomainId = ucloud.String(d.config.DomainId)
	updateUcdnDomainHttpsConfigV2Req.HttpsStatusCn = ucloud.String(getUcdnDomainConfigResp.DomainList[0].HttpsStatusCn)
	updateUcdnDomainHttpsConfigV2Req.HttpsStatusAbroad = ucloud.String(getUcdnDomainConfigResp.DomainList[0].HttpsStatusAbroad)
	updateUcdnDomainHttpsConfigV2Req.HttpsStatusAbroad = ucloud.String(getUcdnDomainConfigResp.DomainList[0].HttpsStatusAbroad)
	updateUcdnDomainHttpsConfigV2Req.CertId = ucloud.Int(certId)
	updateUcdnDomainHttpsConfigV2Req.CertName = ucloud.String(upres.CertName)
	updateUcdnDomainHttpsConfigV2Req.CertType = ucloud.String("ussl")
	updateUcdnDomainHttpsConfigV2Resp, err := d.sdkClient.UpdateUcdnDomainHttpsConfigV2(updateUcdnDomainHttpsConfigV2Req)
	d.logger.Debug("sdk request 'ucdn.UpdateUcdnDomainHttpsConfigV2'", slog.Any("request", updateUcdnDomainHttpsConfigV2Req), slog.Any("response", updateUcdnDomainHttpsConfigV2Resp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'ucdn.UpdateUcdnDomainHttpsConfigV2': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(privateKey, publicKey, projectId string) (*ucloudsdk.UCDNClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}
