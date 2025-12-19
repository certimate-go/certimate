package lecdn

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	leclientsdkv3 "github.com/certimate-go/certimate/pkg/sdk3rd/lecdn/v3/client"
	lemastersdkv3 "github.com/certimate-go/certimate/pkg/sdk3rd/lecdn/v3/master"
)

type DeployerConfig struct {
	// LeCDN 服务地址。
	ServerUrl string `json:"serverUrl"`
	// LeCDN 版本。
	// 可取值 "v3"。
	ApiVersion string `json:"apiVersion"`
	// LeCDN 用户角色。
	// 可取值 "client"、"master"。
	ApiRole string `json:"apiRole"`
	// LeCDN 用户名。
	Username string `json:"username"`
	// LeCDN 用户密码。
	Password string `json:"password"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署资源类型。
	ResourceType string `json:"resourceType"`
	// 证书 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
	// 客户 ID。
	// 部署资源类型为 [RESOURCE_TYPE_CERTIFICATE] 时选填。
	ClientId int64 `json:"clientId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient any
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiVersion, config.ApiRole, config.Username, config.Password, config.AllowInsecureConnections)
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
	if d.config.CertificateId == 0 {
		return errors.New("config `certificateId` is required")
	}

	// 修改证书
	// REF: https://wdk0pwf8ul.feishu.cn/wiki/YE1XwCRIHiLYeKkPupgcXrlgnDd
	switch sdkClient := d.sdkClient.(type) {
	case *leclientsdkv3.Client:
		{
			updateSSLCertReq := &leclientsdkv3.UpdateCertificateRequest{
				Name:        fmt.Sprintf("certimate-%d", time.Now().UnixMilli()),
				Description: "upload from certimate",
				Type:        "upload",
				SSLPEM:      certPEM,
				SSLKey:      privkeyPEM,
				AutoRenewal: false,
			}
			updateSSLCertResp, err := sdkClient.UpdateCertificate(d.config.CertificateId, updateSSLCertReq)
			d.logger.Debug("sdk request 'lecdn.UpdateCertificate'", slog.Int64("certId", d.config.CertificateId), slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'lecdn.UpdateCertificate': %w", err)
			}
		}

	case *lemastersdkv3.Client:
		{
			updateSSLCertReq := &lemastersdkv3.UpdateCertificateRequest{
				ClientId:    d.config.ClientId,
				Name:        fmt.Sprintf("certimate-%d", time.Now().UnixMilli()),
				Description: "upload from certimate",
				Type:        "upload",
				SSLPEM:      certPEM,
				SSLKey:      privkeyPEM,
				AutoRenewal: false,
			}
			updateSSLCertResp, err := sdkClient.UpdateCertificate(d.config.CertificateId, updateSSLCertReq)
			d.logger.Debug("sdk request 'lecdn.UpdateCertificate'", slog.Int64("certId", d.config.CertificateId), slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'lecdn.UpdateCertificate': %w", err)
			}
		}

	default:
		panic("unreachable")
	}

	return nil
}

const (
	sdkVersionV3 = "v3"

	sdkRoleClient = "client"
	sdkRoleMaster = "master"
)

func createSDKClient(serverUrl, apiVersion, apiRole, username, password string, skipTlsVerify bool) (any, error) {
	if apiVersion == sdkVersionV3 && apiRole == sdkRoleClient {
		// v3 版客户端
		client, err := leclientsdkv3.NewClient(serverUrl, username, password)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	} else if apiVersion == sdkVersionV3 && apiRole == sdkRoleMaster {
		// v3 版主控端
		client, err := lemastersdkv3.NewClient(serverUrl, username, password)
		if err != nil {
			return nil, err
		}

		if skipTlsVerify {
			client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
		}

		return client, nil
	}

	return nil, errors.New("lecdn: invalid api version or user role")
}
