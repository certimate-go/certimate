package synologydsm

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	dsmsdk "github.com/certimate-go/certimate/pkg/sdk3rd/synologydsm"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xwait "github.com/certimate-go/certimate/pkg/utils/wait"
)

type DeployerConfig struct {
	// 群晖 DSM 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 群晖 DSM 用户名。
	Username string `json:"username"`
	// 群晖 DSM 用户密码。
	Password string `json:"password"`
	// 群晖 DSM 2FA TOTP 密钥。
	TotpSecret string `json:"totpSecret,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
	// 证书 ID 或描述。
	// 选填。零值时表示新建证书；否则表示更新证书。
	CertificateIdOrDescription string `json:"certificateIdOrDesc,omitempty"`
	// 是否设为默认证书。
	IsDefault bool `json:"isDefault,omitempty"`
}

type Deployer struct {
	config    *DeployerConfig
	logger    *slog.Logger
	sdkClient *dsmsdk.Client
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
	}

	client, err := createSDKClient(config.ServerUrl, config.AllowInsecureConnections)
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
		d.logger = slog.Default()
	} else {
		d.logger = logger
	}
}

func (d *Deployer) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*deployer.DeployResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, intermediateCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 如果启用了 TOTP，则等到下一个时间窗口后生成 OTP 动态密码
	var otpCode string
	if d.config.TotpSecret != "" {
		now := time.Now()
		wait := time.Duration(30-now.Unix()%30) * time.Second
		if wait > 0 {
			wait = wait + 1*time.Second
			d.logger.Info("waiting for the next TOTP time step ...", slog.Int("wait", int(wait.Seconds())))
			xwait.DelayWithContext(ctx, wait)
		}

		now = time.Now()
		otpCodeStr, err := totp.GenerateCode(d.config.TotpSecret, now)
		if err != nil {
			return nil, fmt.Errorf("failed to generate TOTP code: %w", err)
		}

		otpCode = otpCodeStr
	}

	// 登录到群晖 DSM
	loginReq := &dsmsdk.LoginRequest{
		Account:  d.config.Username,
		Password: d.config.Password,
		OtpCode:  otpCode,
	}
	loginResp, err := d.sdkClient.Login(loginReq)
	d.logger.Debug("sdk request 'SYNO.API.Auth:login'", slog.Any("request", loginReq), slog.Any("response", loginResp))
	if err != nil {
		return nil, fmt.Errorf("failed to execute sdk request 'SYNO.API.Auth:login': %w", err)
	}
	defer func() {
		logoutResp, _ := d.sdkClient.Logout()
		d.logger.Debug("sdk request 'SYNO.API.Auth:logout'", slog.Any("response", logoutResp))
	}()

	// 如果原证书 ID 或描述为空，则创建证书；否则更新证书。
	if d.config.CertificateIdOrDescription == "" {
		// 导入证书
		importCertificateReq := &dsmsdk.ImportCertificateRequest{
			ID:          "",
			Description: fmt.Sprintf("certimate-%d", time.Now().UnixMilli()),
			Key:         privkeyPEM,
			Cert:        serverCertPEM,
			InterCert:   intermediateCertPEM,
			AsDefault:   d.config.IsDefault,
		}
		importCertificateResp, err := d.sdkClient.ImportCertificate(importCertificateReq)
		d.logger.Debug("sdk request 'SYNO.Core.Certificate:import'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'SYNO.Core.Certificate:import': %w", err)
		}
	} else {
		// 查找证书列表，找到已有证书
		var certInfo *dsmsdk.CertificateInfo
		listCertificatesResp, err := d.sdkClient.ListCertificates()
		d.logger.Debug("sdk request 'SYNO.Core.Certificate.CRT:list'", slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'SYNO.Core.Certificate.CRT:list': %w", err)
		} else {
			matchedCerts := lo.Filter(listCertificatesResp.Data.Certificates, func(certItem *dsmsdk.CertificateInfo, _ int) bool {
				return certItem.ID == d.config.CertificateIdOrDescription
			})
			if len(matchedCerts) == 0 {
				matchedCerts = lo.Filter(listCertificatesResp.Data.Certificates, func(certItem *dsmsdk.CertificateInfo, _ int) bool {
					return certItem.Description == d.config.CertificateIdOrDescription
				})
			}
			if len(matchedCerts) == 0 {
				return nil, fmt.Errorf("could not find certificate '%s'", d.config.CertificateIdOrDescription)
			} else {
				if len(matchedCerts) > 1 {
					d.logger.Warn("found several certificates matched '%s', using the first one")
				}
				certInfo = matchedCerts[0]
			}
		}

		// 导入证书
		importCertificateReq := &dsmsdk.ImportCertificateRequest{
			ID:          certInfo.ID,
			Description: certInfo.Description,
			Key:         privkeyPEM,
			Cert:        serverCertPEM,
			InterCert:   intermediateCertPEM,
			AsDefault:   d.config.IsDefault || certInfo.IsDefault,
		}
		importCertificateResp, err := d.sdkClient.ImportCertificate(importCertificateReq)
		d.logger.Debug("sdk request 'SYNO.Core.Certificate:import'", slog.Any("request", importCertificateReq), slog.Any("response", importCertificateResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'SYNO.Core.Certificate:import': %w", err)
		}
	}

	if d.config.IsDefault {
		// 查找证书列表，找到默认证书
		listCertificatesResp, err := d.sdkClient.ListCertificates()
		d.logger.Debug("sdk request 'SYNO.Core.Certificate.CRT:list'", slog.Any("response", listCertificatesResp))
		if err != nil {
			return nil, fmt.Errorf("failed to execute sdk request 'SYNO.Core.Certificate.CRT:list': %w", err)
		} else {
			var defaultCertId string
			for _, certItem := range listCertificatesResp.Data.Certificates {
				if certItem.IsDefault {
					defaultCertId = certItem.ID
					break
				}
			}

			if defaultCertId != "" {
				settings := make([]*dsmsdk.ServiceCertificateSetting, 0)
				for _, certItem := range listCertificatesResp.Data.Certificates {
					if certItem.ID == defaultCertId {
						continue
					}

					for _, service := range certItem.Services {
						settings = append(settings, &dsmsdk.ServiceCertificateSetting{
							Service:   service,
							CertID:    defaultCertId,
							OldCertID: certItem.ID,
						})
					}
				}

				// 应用到所有服务并重启
				if len(settings) > 0 {
					setServiceCertificateReq := &dsmsdk.SetServiceCertificateRequest{
						Settings: settings,
					}
					setServiceCertificateResp, err := d.sdkClient.SetServiceCertificate(setServiceCertificateReq)
					d.logger.Debug("sdk request 'SYNO.Core.Certificate.Service:set'", slog.Any("request", setServiceCertificateReq), slog.Any("response", setServiceCertificateResp))
					if err != nil {
						return nil, fmt.Errorf("failed to execute sdk request 'SYNO.Core.Certificate.Service:set': %w", err)
					}
				}
			}
		}
	}

	return &deployer.DeployResult{}, nil
}

func createSDKClient(serverUrl string, skipTlsVerify bool) (*dsmsdk.Client, error) {
	client, err := dsmsdk.NewClient(serverUrl)
	if err != nil {
		return nil, err
	}

	if skipTlsVerify {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return client, nil
}
