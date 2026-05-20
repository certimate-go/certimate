package ssh

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/certimate-go/certimate/internal/tools/ssh"
	"github.com/certimate-go/certimate/pkg/core/deployer"
	"github.com/certimate-go/certimate/pkg/core/deployer/providers/local/shared"
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
	// 证书格式。
	FileFormat string `json:"fileFormat"`
	// 私钥文件路径。
	FilePathForKey string `json:"filePathForKey,omitempty"`
	// 证书文件路径。
	FilePathForCrt string `json:"filePathForCrt,omitempty"`
	// 证书文件（仅含服务器证书）路径。
	// 选填。
	FilePathForCrtOnlyServer string `json:"filePathForCrtOnlyServer,omitempty"`
	// 证书文件（仅含中间证书）路径。
	// 选填。
	FilePathForCrtOnlyIntermedia string `json:"filePathForCrtOnlyIntermedia,omitempty"`
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
}

type Deployer struct {
	config *DeployerConfig
	logger *slog.Logger
}

var _ deployer.Provider = (*Deployer)(nil)

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

func (d *Deployer) Deploy(ctx context.Context, certPEM, privkeyPEM string) (*deployer.DeployResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 连接到 SSH
	sshClient, err := createSshClient(*d.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshClient.Close()
	d.logger.Info("ssh connected")

	// 执行前置命令
	if d.config.PreCommand != "" {
		command := d.config.PreCommand
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}", d.config.FilePathForCrt)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_SERVER_PATH}", d.config.FilePathForCrtOnlyServer)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_INTERMEDIA_PATH}", d.config.FilePathForCrtOnlyIntermedia)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}", d.config.FilePathForKey)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}", d.config.PfxPassword)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_ALIAS}", d.config.JksAlias)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_KEYPASS}", d.config.JksKeypass)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_STOREPASS}", d.config.JksStorepass)

		stdout, stderr, err := xssh.RunCommand(sshClient.RawClient(), command)
		d.logger.Debug("run pre-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute pre-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	// 上传证书和私钥文件
	switch d.config.FileFormat {
	case FILE_FORMAT_PEM:
		{
			if d.config.FilePathForKey != "" {
				if err := xssh.WriteRemoteString(sshClient.RawClient(), d.config.FilePathForKey, privkeyPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload private key file: %w", err)
				}
				d.logger.Info("ssl private key file uploaded", slog.String("path", d.config.FilePathForKey))
			}

			if d.config.FilePathForCrt != "" {
				if err := xssh.WriteRemoteString(sshClient.RawClient(), d.config.FilePathForCrt, certPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.FilePathForCrt))
			}

			if d.config.FilePathForCrtOnlyServer != "" {
				if err := xssh.WriteRemoteString(sshClient.RawClient(), d.config.FilePathForCrtOnlyServer, serverCertPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to save server certificate file: %w", err)
				}
				d.logger.Info("ssl server certificate file uploaded", slog.String("path", d.config.FilePathForCrtOnlyServer))
			}

			if d.config.FilePathForCrtOnlyIntermedia != "" {
				if err := xssh.WriteRemoteString(sshClient.RawClient(), d.config.FilePathForCrtOnlyIntermedia, intermediaCertPEM, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to save intermedia certificate file: %w", err)
				}
				d.logger.Info("ssl intermedia certificate file uploaded", slog.String("path", d.config.FilePathForCrtOnlyIntermedia))
			}
		}

	case FILE_FORMAT_PFX:
		{
			if d.config.PfxPassword == "" {
				return nil, fmt.Errorf("config `pfxPassword` is required")
			}

			pfxEncoder, err := shared.ResolvePfxEncoder(d.config.PfxEncoder)
			if err != nil {
				return nil, fmt.Errorf("config `pfxEncoder` is invalid: %w", err)
			}

			pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword, pfxEncoder)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
			}
			d.logger.Info("ssl certificate transformed to pfx")

			if d.config.FilePathForCrt != "" {
				if err := xssh.WriteRemote(sshClient.RawClient(), d.config.FilePathForCrt, pfxData, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.FilePathForCrt))
			}
		}

	case FILE_FORMAT_JKS:
		{
			if d.config.JksAlias == "" {
				return nil, fmt.Errorf("config `jksAlias` is required")
			}
			if d.config.JksKeypass == "" {
				return nil, fmt.Errorf("config `jksKeypass` is required")
			}
			if d.config.JksStorepass == "" {
				return nil, fmt.Errorf("config `jksStorepass` is required")
			}

			jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
			if err != nil {
				return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
			}
			d.logger.Info("ssl certificate transformed to jks")

			if d.config.FilePathForCrt != "" {
				if err := xssh.WriteRemote(sshClient.RawClient(), d.config.FilePathForCrt, jksData, d.config.UseSCP); err != nil {
					return nil, fmt.Errorf("failed to upload certificate file: %w", err)
				}
				d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.FilePathForCrt))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported file format '%s'", d.config.FileFormat)
	}

	// 执行后置命令
	if d.config.PostCommand != "" {
		command := d.config.PostCommand
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}", d.config.FilePathForCrt)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_SERVER_PATH}", d.config.FilePathForCrtOnlyServer)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_INTERMEDIA_PATH}", d.config.FilePathForCrtOnlyIntermedia)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}", d.config.FilePathForKey)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}", d.config.PfxPassword)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_ALIAS}", d.config.JksAlias)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_KEYPASS}", d.config.JksKeypass)
		command = strings.ReplaceAll(command, "${CERTIMATE_DEPLOYER_CMDVAR_JKS_STOREPASS}", d.config.JksStorepass)

		stdout, stderr, err := xssh.RunCommand(sshClient.RawClient(), command)
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
