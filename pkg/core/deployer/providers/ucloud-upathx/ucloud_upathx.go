package ucloudupathx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/core/certmgr"
	mcertmgr "github.com/certimate-go/certimate/pkg/core/certmgr/providers/ucloud-upathx"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/upathx"
)

type DeployerConfig struct {
	// 优刻得 API 私钥。
	PrivateKey string `json:"privateKey"`
	// 优刻得 API 公钥。
	PublicKey string `json:"publicKey"`
	// 优刻得项目 ID。
	ProjectId string `json:"projectId,omitempty"`
	// 加速器实例 ID。
	AcceleratorId string `json:"acceleratorId"`
	// 加速器监听端口。
	ListenerPort int32 `json:"listenerPort"`
}

type Deployer struct {
	config     *DeployerConfig
	logger     *slog.Logger
	sdkClient  *ucloudsdk.UPathXClient
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
	if d.config.AcceleratorId == "" {
		return nil, errors.New("config `acceleratorId` is required")
	}
	if d.config.ListenerPort == 0 {
		return nil, errors.New("config `listenerPort` is required")
	}

	// 上传证书
	upres, err := d.sdkCertmgr.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", slog.Any("result", upres))
	}

	// 绑定 PathX SSL 证书
	// REF: https://docs.ucloud.cn/api/pathx-api/bind_path_xssl
	bindPathXSSLReq := d.sdkClient.NewBindPathXSSLRequest()
	bindPathXSSLReq.UGAId = ucloud.String(d.config.AcceleratorId)
	bindPathXSSLReq.Port = []int{int(d.config.ListenerPort)}
	bindPathXSSLReq.SSLId = ucloud.String(upres.CertId)
	bindPathXSSLResp, err := d.sdkClient.BindPathXSSL(bindPathXSSLReq)
	d.logger.Debug("sdk request 'pathx.BindPathXSSL'", slog.Any("request", bindPathXSSLReq), slog.Any("response", bindPathXSSLResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'pathx.BindPathXSSL': %w", err)
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(privateKey, publicKey, projectId string) (*ucloudsdk.UPathXClient, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("ucloud: invalid private key")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("ucloud: invalid public key")
	}

	cfg := ucloud.NewConfig()
	cfg.ProjectId = projectId

	// PathX 相关接口要求必传 ProjectId 参数
	if cfg.ProjectId == "" {
		defaultProjectId, err := getSDKDefaultProjectId(privateKey, publicKey)
		if err != nil {
			return nil, err
		}

		cfg.ProjectId = defaultProjectId
	}

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := ucloudsdk.NewClient(&cfg, &credential)
	return client, nil
}

func getSDKDefaultProjectId(privateKey, publicKey string) (string, error) {
	cfg := ucloud.NewConfig()

	credential := auth.NewCredential()
	credential.PrivateKey = privateKey
	credential.PublicKey = publicKey

	client := uaccount.NewClient(&cfg, &credential)

	request := client.NewGetProjectListRequest()
	response, err := client.GetProjectList(request)
	if err != nil {
		return "", err
	}

	for _, projectItem := range response.ProjectSet {
		if projectItem.IsDefault {
			return projectItem.ProjectId, nil
		}
	}

	return "", errors.New("ucloud: no default project found")
}
