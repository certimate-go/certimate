package huaweiibmc

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/certimate-go/certimate/pkg/core"
	ibmcsdk "github.com/certimate-go/certimate/pkg/sdk3rd/huaweiibmc"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xloop "github.com/certimate-go/certimate/pkg/utils/loop"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// iBMC 主机。
	Host string `json:"host"`
	// iBMC 用户名。
	Username string `json:"username"`
	// iBMC 密码。
	Password string `json:"password"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 是否自动重启。
	AutoRestart bool `json:"autoRestart,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the iBMC deployer is nil")
	}

	return &Deployer{
		config: config,
		logger: slog.Default(),
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
	// 转换证书格式
	certPFXPwd := make([]byte, 24)
	if _, err := rand.Read(certPFXPwd); err != nil {
		return nil, fmt.Errorf("failed to generate PFX password: %w", err)
	}
	certPFXPwdHex := hex.EncodeToString(certPFXPwd)
	certPFX, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, certPFXPwdHex, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to transform certificate from PEM to PFX: %w", err)
	}

	// 创建 iBMC 客户端
	client, err := createSDKClient(d.config.Host, d.config.Username, d.config.Password, d.config.AllowInsecureConnections)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	// 创建会话
	createSessionResp, err := client.CreateSessionWithContext(ctx)
	d.logger.Debug("sdk request 'SessionService.CreateSession'", slog.Any("response", createSessionResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'SessionService.CreateSession': %w", err)
	} else {
		defer client.DeleteSessionWithContext(ctx)
	}

	// 查询管理集合资源信息
	listManagersResp, err := client.ListManagersWithContext(ctx)
	d.logger.Debug("sdk request 'Managers.ListManagers'", slog.Any("response", listManagersResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'Managers.ListManagers': %w", err)
	}

	// 批量更新管理证书
	if len(listManagersResp.Members) == 0 {
		d.logger.Info("no ibmc managers to deploy")
	} else {
		d.logger.Info("found ibmc managers to deploy", slog.Any("managers", listManagersResp.Members))

		if err := xloop.ForRangeAllWithContext(ctx, listManagersResp.Members, func(ctx context.Context, managerInfo *ibmcsdk.Entity, _ int) error {
			importCustomCertificateReq := &ibmcsdk.ImportCustomCertificateToManagerRequest{
				ManagerID:       managerInfo.ID,
				ManagerLocation: managerInfo.ODataID,
				Certificate:     base64.StdEncoding.EncodeToString(certPFX),
				Password:        certPFXPwdHex,
			}
			importCustomCertificateResp, err := client.ImportCustomCertificateToManagerWithContext(ctx, importCustomCertificateReq)
			d.logger.Debug("sdk request 'Managers.SecurityService.HttpsCert.ImportCustomCertificateToManager'", slog.String("params.managerODataId", managerInfo.ODataID), slog.Any("request", importCustomCertificateReq), slog.Any("response", importCustomCertificateResp))
			if err != nil {
				return fmt.Errorf("failed to execute sdk request 'Managers.SecurityService.HttpsCert.ImportCustomCertificateToManager': %w", err)
			}

			if d.config.AutoRestart {
				resetManagerReq := &ibmcsdk.ResetManagerRequest{
					ManagerID:       managerInfo.ID,
					ManagerLocation: managerInfo.ODataID,
					ResetType:       "ForceRestart",
				}
				resetManagerResp, err := client.ResetManagerWithContext(ctx, resetManagerReq)
				d.logger.Debug("sdk request 'Managers.ResetManager'", slog.String("params.managerODataId", managerInfo.ODataID), slog.Any("request", resetManagerReq), slog.Any("response", resetManagerResp))
				if err != nil {
					return fmt.Errorf("failed to execute sdk request 'Managers.ResetManager': %w", err)
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &DeployResult{}, nil
}

func createSDKClient(host, username, password string, skipTlsVerify bool) (*ibmcsdk.Client, error) {
	serverUrl := ""
	if strings.Contains(host, "://") {
		serverUrl = host
	} else {
		if net.ParseIP(host) != nil && strings.Contains(host, ":") {
			host = "[" + host + "]"
		}
		serverUrl = "https://" + host
	}

	client, err := ibmcsdk.NewClient(
		serverUrl,
		ibmcsdk.WithLogins(username, password),
	)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}

func generateCertificatePassword() (string, error) {
	password := make([]byte, 24)
	if _, err := rand.Read(password); err != nil {
		return "", err
	}
	return hex.EncodeToString(password), nil
}
