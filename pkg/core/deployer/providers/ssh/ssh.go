package ssh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/internal/tools/ssh"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xssh "github.com/certimate-go/certimate/pkg/utils/ssh"
)

type ServerConfig struct {
	// SSH 主机。
	SshHost string `json:"sshHost"`
	// SSH 端口。
	// 零值时默认值 22。
	SshPort int32 `json:"sshPort,omitempty"`
	// SSH 认证方式。
	// 可取值 "none"、"password"、"key"。
	// 零值时根据有无密码或私钥字段决定。
	SshAuthMethod string `json:"sshAuthMethod,omitempty"`
	// SSH 登录用户名。
	// 零值时默认值 "root"。
	SshUsername string `json:"sshUsername,omitempty"`
	// SSH 登录密码。
	SshPassword string `json:"sshPassword,omitempty"`
	// SSH 登录私钥。
	SshKey string `json:"sshKey,omitempty"`
	// SSH 登录私钥口令。
	SshKeyPassphrase string `json:"sshKeyPassphrase,omitempty"`
}

type DeployerConfig struct {
	ServerConfig

	// 跳板机配置数组。
	JumpServers []ServerConfig `json:"jumpServers,omitempty"`
	// 是否回退使用 SCP。
	UseSCP bool `json:"useSCP,omitempty"`
	// 前置命令。
	PreCommand string `json:"preCommand,omitempty"`
	// 后置命令。
	PostCommand string `json:"postCommand,omitempty"`
	// 输出证书格式。
	OutputFormat string `json:"outputFormat,omitempty"`
	// 输出私钥文件路径。
	OutputKeyPath string `json:"outputKeyPath,omitempty"`
	// 输出证书文件路径。
	OutputCertPath string `json:"outputCertPath,omitempty"`
	// 输出服务器证书文件路径。
	// 选填。
	OutputServerCertPath string `json:"outputServerCertPath,omitempty"`
	// 输出中间证书文件路径。
	// 选填。
	OutputIntermediaCertPath string `json:"outputIntermediaCertPath,omitempty"`
	// PFX 导出密码。
	// 证书格式为 PFX 时必填。
	PfxPassword string `json:"pfxPassword,omitempty"`
	// JKS 别名。
	// 证书格式为 JKS 时必填。
	JksAlias string `json:"jksAlias,omitempty"`
	// JKS 密钥密码。
	// 证书格式为 JKS 时必填。
	JksKeypass string `json:"jksKeypass,omitempty"`
	// JKS 存储密码。
	// 证书格式为 JKS 时必填。
	JksStorepass string `json:"jksStorepass,omitempty"`
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

func NewDeployer(config *DeployerConfig) (*Deployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the deployer provider is nil")
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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	client, err := createSshClient(*d.config)
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to create SSH client: %w", err)
	}

	d.logger.Info("ssh connected")
	defer client.Close()

	// 执行前置命令
	if d.config.PreCommand != "" {
		command := d.config.PreCommand
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}", d.config.OutputCertPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_SERVER_PATH}", d.config.OutputServerCertPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_INTERMEDIA_PATH}", d.config.OutputIntermediaCertPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}", d.config.OutputKeyPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}", d.config.PfxPassword)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_ALIAS}", d.config.JksAlias)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_KEYPASS}", d.config.JksKeypass)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_STOREPASS}", d.config.JksStorepass)

		stdout, stderr, err := xssh.RunCommand(client.RawClient(), command)
		d.logger.Debug("run pre-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute pre-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	// 上传证书和私钥文件
	switch d.config.OutputFormat {
	case OUTPUT_FORMAT_PEM:
		{
			if d.config.OutputKeyPath != "" {
				if err := xssh.WriteRemoteString(client.RawClient(), d.config.OutputKeyPath, privkeyPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				d.logger.Info("ssl private key file uploaded", slog.String("path", d.config.OutputKeyPath))
			}

			if d.config.OutputCertPath != "" {
				if err := xssh.WriteRemoteString(client.RawClient(), d.config.OutputCertPath, certPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))
			}

			if d.config.OutputServerCertPath != "" {
				if err := xssh.WriteRemoteString(client.RawClient(), d.config.OutputServerCertPath, serverCertPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to save server certificate file: %w", err)
				}
				d.logger.Info("ssl server certificate file uploaded", slog.String("path", d.config.OutputServerCertPath))
			}

			if d.config.OutputIntermediaCertPath != "" {
				if err := xssh.WriteRemoteString(client.RawClient(), d.config.OutputIntermediaCertPath, intermediaCertPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to save intermedia certificate file: %w", err)
				}
				d.logger.Info("ssl intermedia certificate file uploaded", slog.String("path", d.config.OutputIntermediaCertPath))
			}
		}

	case OUTPUT_FORMAT_PFX:
		{
			pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
			}
			d.logger.Info("ssl certificate transformed to pfx")

			if d.config.OutputCertPath != "" {
				if err := xssh.WriteRemote(client.RawClient(), d.config.OutputCertPath, pfxData, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))
			}
		}

	case OUTPUT_FORMAT_JKS:
		{
			jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
			}
			d.logger.Info("ssl certificate transformed to jks")

			if d.config.OutputCertPath != "" {
				if err := xssh.WriteRemote(client.RawClient(), d.config.OutputCertPath, jksData, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported output format '%s'", d.config.OutputFormat)
	}

	// 执行后置命令
	if d.config.PostCommand != "" {
		command := d.config.PostCommand
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}", d.config.OutputCertPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_SERVER_PATH}", d.config.OutputServerCertPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_INTERMEDIA_PATH}", d.config.OutputIntermediaCertPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}", d.config.OutputKeyPath)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}", d.config.PfxPassword)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_ALIAS}", d.config.JksAlias)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_KEYPASS}", d.config.JksKeypass)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_STOREPASS}", d.config.JksStorepass)

		stdout, stderr, err := xssh.RunCommand(client.RawClient(), command)
		d.logger.Debug("run post-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute post-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	return &deployer.DeployResult{}, nil
}

func createSshClient(config DeployerConfig) (*ssh.Client, error) {
	clientCfg := ssh.NewDefaultConfig()
	clientCfg.Host = config.SshHost
	clientCfg.Port = int(config.SshPort)
	clientCfg.AuthMethod = ssh.AuthMethodType(config.SshAuthMethod)
	clientCfg.Username = config.SshUsername
	clientCfg.Password = config.SshPassword
	clientCfg.Key = config.SshKey
	clientCfg.KeyPassphrase = config.SshKeyPassphrase
	for _, jumpServer := range config.JumpServers {
		jumpServerCfg := ssh.NewServerConfig()
		jumpServerCfg.Host = jumpServer.SshHost
		jumpServerCfg.Port = int(jumpServer.SshPort)
		jumpServerCfg.AuthMethod = ssh.AuthMethodType(jumpServer.SshAuthMethod)
		jumpServerCfg.Username = jumpServer.SshUsername
		jumpServerCfg.Password = jumpServer.SshPassword
		jumpServerCfg.Key = jumpServer.SshKey
		jumpServerCfg.KeyPassphrase = jumpServer.SshKeyPassphrase
		clientCfg.JumpServers = append(clientCfg.JumpServers, *jumpServerCfg)
	}

	client, err := ssh.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
