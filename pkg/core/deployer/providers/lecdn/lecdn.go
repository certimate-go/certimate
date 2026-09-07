package lecdn

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
	lecdnclientsdkv3 "github.com/certimate-go/certimate/pkg/sdk3rd/lecdn/client/v3"
	lecdnclientsdkv4 "github.com/certimate-go/certimate/pkg/sdk3rd/lecdn/client/v4"
	lecdnmastersdkv3 "github.com/certimate-go/certimate/pkg/sdk3rd/lecdn/master/v3"
	lecdnmastersdkv4 "github.com/certimate-go/certimate/pkg/sdk3rd/lecdn/master/v4"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// LeCDN 服务地址。
	ServerUrl string `json:"serverUrl"`
	// LeCDN 版本。
	// 可取值 "v3"、"v4"。
	ApiVersion string `json:"apiVersion"`
	// LeCDN 用户角色。
	// 可取值 "client"、"master"。
	ApiRole string `json:"apiRole"`
	// LeCDN API 认证方式。
	// 可取值 "password"、"apikey"。
	// 零值时默认值 [AUTH_METHOD_PASSWORD]。
	AuthMethod string `json:"authMethod,omitempty"`
	// LeCDN 用户名。
	Username string `json:"username,omitempty"`
	// LeCDN 用户密码。
	Password string `json:"password,omitempty"`
	// LeCDN API Key。
	ApiKey string `json:"apiKey,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 部署目标。
	DeployTarget string `json:"deployTarget"`
	// 证书 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时必填。
	CertificateId int64 `json:"certificateId,omitempty"`
	// 客户 ID。
	// 部署目标为 [DEPLOY_TARGET_CERTIFICATE] 时选填。
	ClientId int64 `json:"clientId,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient any
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.ApiVersion, config.ApiRole, config.AuthMethod, config.Username, config.Password, config.ApiKey, config.AllowInsecureConnections)
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*DeployResult, error) {
	// 根据部署目标决定业务流程
	switch d.config.DeployTarget {
	case DEPLOY_TARGET_CERTIFICATE:
		if err := d.deployToCertificate(ctx, certPEM, privkeyPEM); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported deploy target '%s'", d.config.DeployTarget)
	}

	return &DeployResult{}, nil
}

func (d *Deployer) deployToCertificate(ctx context.Context, certPEM, privkeyPEM string) error {
	if d.config.CertificateId == 0 {
		return fmt.Errorf("config `certificateId` is required")
	}

	// 生成新证书名（需符合 LeCDN 命名规则）
	certName := fmt.Sprintf("certimate-%d", time.Now().UnixMilli())
	certDesc := "upload from Certimate"

	// 修改证书
	// REF: https://wdk0pwf8ul.feishu.cn/wiki/YE1XwCRIHiLYeKkPupgcXrlgnDd
	switch sdkClient := d.sdkClient.(type) {
	case *lecdnclientsdkv3.Client:
		{
			updateSSLCertReq := &lecdnclientsdkv3.UpdateCertificateRequest{
				Name:        certName,
				Description: certDesc,
				Type:        "upload",
				SSLPEM:      certPEM,
				SSLKey:      privkeyPEM,
				AutoRenewal: false,
			}
			updateSSLCertResp, err := sdkClient.UpdateCertificateWithContext(ctx, d.config.CertificateId, updateSSLCertReq)
			d.logger.Debug("sdk request 'UpdateCertificate'", slog.Int64("params.certId", d.config.CertificateId), slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'UpdateCertificate': %w", err)
			}
		}

	case *lecdnmastersdkv3.Client:
		{
			updateSSLCertReq := &lecdnmastersdkv3.UpdateCertificateRequest{
				ClientId:    d.config.ClientId,
				Name:        certName,
				Description: certDesc,
				Type:        "upload",
				SSLPEM:      certPEM,
				SSLKey:      privkeyPEM,
				AutoRenewal: false,
			}
			updateSSLCertResp, err := sdkClient.UpdateCertificateWithContext(ctx, d.config.CertificateId, updateSSLCertReq)
			d.logger.Debug("sdk request 'UpdateCertificate'", slog.Int64("params.certId", d.config.CertificateId), slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'UpdateCertificate': %w", err)
			}
		}

	case *lecdnclientsdkv4.Client:
		{
			updateSSLCertReq := &lecdnclientsdkv4.UpdateCertificateRequest{
				Name:        certName,
				Description: certDesc,
				SSLPEM:      certPEM,
				SSLKey:      privkeyPEM,
				AutoRenewal: false,
			}
			updateSSLCertResp, err := sdkClient.UpdateCertificateWithContext(ctx, d.config.CertificateId, updateSSLCertReq)
			d.logger.Debug("sdk request 'UpdateCertificate'", slog.Int64("params.certId", d.config.CertificateId), slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'UpdateCertificate': %w", err)
			}
		}

	case *lecdnmastersdkv4.Client:
		{
			updateSSLCertReq := &lecdnmastersdkv4.UpdateCertificateRequest{
				Name:        certName,
				Description: certDesc,
				SSLPEM:      certPEM,
				SSLKey:      privkeyPEM,
				AutoRenewal: false,
			}
			updateSSLCertResp, err := sdkClient.UpdateCertificateWithContext(ctx, d.config.CertificateId, updateSSLCertReq)
			d.logger.Debug("sdk request 'UpdateCertificate'", slog.Int64("params.certId", d.config.CertificateId), slog.Any("request", updateSSLCertReq), slog.Any("response", updateSSLCertResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'UpdateCertificate': %w", err)
			}
		}

	default:
		panic("unreachable")
	}

	return nil
}

const (
	sdkVersionV3 = "v3"
	sdkVersionV4 = "v4"

	sdkRoleClient = "client"
	sdkRoleMaster = "master"
)

func createSDKClient(serverUrl, apiVersion, apiRole, authMethod, username, password, apiKey string, skipTlsVerify bool) (any, error) {
	switch apiVersion {
	case sdkVersionV3:
		{
			switch apiRole {
			case sdkRoleClient:
				{
					// v3 版客户端
					var client *lecdnclientsdkv3.Client
					var err error

					switch authMethod {
					case "", AUTH_METHOD_PASSWORD:
						client, err = lecdnclientsdkv3.NewClient(serverUrl, lecdnclientsdkv3.WithLogins(username, password))
					case AUTH_METHOD_APIKEY:
						client, err = lecdnclientsdkv3.NewClient(serverUrl, lecdnclientsdkv3.WithLogins(username, password))
					default:
						err = fmt.Errorf("unsupported auth method '%s'", authMethod)
					}

					if err != nil {
						return nil, err
					}

					if skipTlsVerify {
						client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
					}

					return client, nil
				}

			case sdkRoleMaster:
				{
					// v3 版主控端
					var client *lecdnmastersdkv3.Client
					var err error

					switch authMethod {
					case "", AUTH_METHOD_PASSWORD:
						client, err = lecdnmastersdkv3.NewClient(serverUrl, lecdnmastersdkv3.WithLogins(username, password))
					case AUTH_METHOD_APIKEY:
						client, err = lecdnmastersdkv3.NewClient(serverUrl, lecdnmastersdkv3.WithLogins(username, password))
					default:
						err = fmt.Errorf("unsupported auth method '%s'", authMethod)
					}

					if err != nil {
						return nil, err
					}

					if skipTlsVerify {
						client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
					}

					return client, nil
				}
			}
		}

	case sdkVersionV4:
		{
			switch apiRole {
			case sdkRoleClient:
				{
					// v4 版客户端
					var client *lecdnclientsdkv4.Client
					var err error

					switch authMethod {
					case "", AUTH_METHOD_PASSWORD:
						client, err = lecdnclientsdkv4.NewClient(serverUrl, lecdnclientsdkv4.WithLogins(username, password))
					case AUTH_METHOD_APIKEY:
						client, err = lecdnclientsdkv4.NewClient(serverUrl, lecdnclientsdkv4.WithLogins(username, password))
					default:
						err = fmt.Errorf("unsupported auth method '%s'", authMethod)
					}

					if err != nil {
						return nil, err
					}

					if skipTlsVerify {
						client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
					}

					return client, nil
				}

			case sdkRoleMaster:
				{
					// v4 版主控端
					var client *lecdnmastersdkv4.Client
					var err error

					switch authMethod {
					case "", AUTH_METHOD_PASSWORD:
						client, err = lecdnmastersdkv4.NewClient(serverUrl, lecdnmastersdkv4.WithLogins(username, password))
					case AUTH_METHOD_APIKEY:
						client, err = lecdnmastersdkv4.NewClient(serverUrl, lecdnmastersdkv4.WithLogins(username, password))
					default:
						err = fmt.Errorf("unsupported auth method '%s'", authMethod)
					}

					if err != nil {
						return nil, err
					}

					if skipTlsVerify {
						client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
					}

					return client, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("lecdn: invalid api version or user role")
}
