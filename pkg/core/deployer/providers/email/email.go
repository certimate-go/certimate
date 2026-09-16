package email

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/internal/tools/smtp"
	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xcertx509 "github.com/certimate-go/certimate/pkg/utils/cert/x509"
)

type (
	Provider     = core.Deployer
	DeployResult = core.DeployerDeployResult
)

type DeployerConfig struct {
	// SMTP 服务器地址。
	SmtpHost string `json:"smtpHost"`
	// SMTP 服务器端口。
	// 零值时根据是否启用 TLS 决定。
	SmtpPort int32 `json:"smtpPort"`
	// 是否启用 TLS。
	SmtpTls bool `json:"smtpTls"`
	// 用户名。
	Username string `json:"username,omitempty"`
	// 密码。
	Password string `json:"password,omitempty"`
	// 发件人邮箱。
	SenderAddress string `json:"senderAddress"`
	// 发件人显示名称。
	SenderName string `json:"senderName,omitempty"`
	// 收件人邮箱。
	ReceiverAddress string `json:"receiverAddress"`
	// 证书格式。
	// 可取值 [FILE_FORMAT_PEM]、[FILE_FORMAT_PFX]、[FILE_FORMAT_JKS]。
	// 零值时默认值 [FILE_FORMAT_PEM]。
	FileFormat string `json:"fileFormat,omitempty"`
	// PFX 导出密码。
	// 证书格式为 [FILE_FORMAT_PFX] 时必填。
	PfxPassword string `json:"pfxPassword,omitempty"`
	// PFX 编码器。
	// 证书格式为 [FILE_FORMAT_PFX] 时可选。
	PfxEncoder string `json:"pfxEncoder,omitempty"`
	// JKS 别名。
	// 证书格式为 [FILE_FORMAT_JKS] 时必填。
	JksAlias string `json:"jksAlias,omitempty"`
	// JKS 密钥密码。
	// 证书格式为 [FILE_FORMAT_JKS] 时必填。
	JksKeypass string `json:"jksKeypass,omitempty"`
	// JKS 存储密码。
	// 证书格式为 [FILE_FORMAT_JKS] 时必填。
	JksStorepass string `json:"jksStorepass,omitempty"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the deployer provider is nil")
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
	// 解析证书信息（用于 zip 内文件名）
	fileFormat := d.config.FileFormat
	if fileFormat == "" {
		fileFormat = FILE_FORMAT_PEM
	}

	canonicalName := "certificate"
	if cert, err := xcert.ParseCertificateFromPEM(certPEM); err == nil {
		sans := xcertx509.GetSubjectAltNames(cert)
		if len(sans) > 0 {
			canonicalName = strings.ReplaceAll(sans[0], "*", "_")
		}
	}

	// 生成证书压缩包（与页面「下载证书」完全一致）
	zipData, err := xcert.BuildCertificateArchive(certPEM, privkeyPEM, canonicalName, xcert.CertificateArchiveOptions{
		FileFormat:   fileFormat,
		PfxPassword:  d.config.PfxPassword,
		PfxEncoder:   d.config.PfxEncoder,
		JksAlias:     d.config.JksAlias,
		JksKeypass:   d.config.JksKeypass,
		JksStorepass: d.config.JksStorepass,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build certificate archive: %w", err)
	}

	// 解析证书主题（用于邮件标题）
	certSubject := ""
	if cert, err := xcert.ParseCertificateFromPEM(certPEM); err == nil {
		certSubject = cert.Subject.CommonName
	}

	// 发送邮件
	clientCfg := smtp.NewDefaultConfig()
	clientCfg.Host = d.config.SmtpHost
	clientCfg.Port = int(d.config.SmtpPort)
	clientCfg.Username = d.config.Username
	clientCfg.Password = d.config.Password
	clientCfg.UseSsl = d.config.SmtpTls
	clientCfg.SkipTlsVerify = d.config.AllowInsecureConnections
	client, err := smtp.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}

	defer client.Close()

	msg := smtp.NewMessage()
	if certSubject != "" {
		msg.Subject(fmt.Sprintf("SSL 证书部署：%s", certSubject))
	} else {
		msg.Subject("SSL 证书部署")
	}

	if d.config.SenderName == "" {
		if err := msg.From(d.config.SenderAddress); err != nil {
			return nil, fmt.Errorf("failed to set sender address: %w", err)
		}
	} else {
		if err := msg.FromFormat(d.config.SenderName, d.config.SenderAddress); err != nil {
			return nil, fmt.Errorf("failed to set sender address: %w", err)
		}
	}

	receiverAddresses := strings.Split(d.config.ReceiverAddress, ",")
	for _, addr := range receiverAddresses {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		if err := msg.To(addr); err != nil {
			return nil, fmt.Errorf("failed to set receiver address: %w", err)
		}
	}

	msg.SetBodyString(smtp.MIMETypeTextPlain, fmt.Sprintf("附件为部署的 SSL 证书压缩包（格式：%s）。", fileFormat))
	if err := msg.AttachReader(fmt.Sprintf("%s.zip", canonicalName), bytes.NewReader(zipData)); err != nil {
		return nil, fmt.Errorf("failed to attach certificate: %w", err)
	}

	if err := client.Send(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to send mail: %w", err)
	}

	d.logger.Info("ssl certificate sent via email")

	return &DeployResult{}, nil
}
