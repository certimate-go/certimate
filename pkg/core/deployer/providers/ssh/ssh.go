package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/certimate-go/certimate/pkg/core/deployer"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
	xssh "github.com/certimate-go/certimate/pkg/utils/ssh"
)

type ServerConfig struct {
	// SSH 主机。
	// 零值时默认值 "localhost"。
	SshHost string `json:"sshHost,omitempty"`
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
	var err error

	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	// 创建 TCP 链接
	var targetConn net.Conn
	if len(d.config.JumpServers) > 0 {
		var jumpClient *ssh.Client
		for i, jumpServerConf := range d.config.JumpServers {
			d.logger.Info(fmt.Sprintf("connecting to jump server [%d]", i+1), slog.String("host", jumpServerConf.SshHost))

			var jumpConn net.Conn
			// 第一个连接是主机发起，后续通过跳板机发起
			if jumpClient == nil {
				jumpConn, err = net.Dial("tcp", net.JoinHostPort(jumpServerConf.SshHost, strconv.Itoa(int(jumpServerConf.SshPort))))
			} else {
				jumpConn, err = jumpClient.DialContext(ctx, "tcp", net.JoinHostPort(jumpServerConf.SshHost, strconv.Itoa(int(jumpServerConf.SshPort))))
			}
			if err != nil {
				return nil, fmt.Errorf("failed to connect to jump server [%d]: %w", i+1, err)
			}
			defer jumpConn.Close()

			newClient, err := createSshClient(
				jumpConn,
				jumpServerConf.SshHost,
				jumpServerConf.SshPort,
				jumpServerConf.SshAuthMethod,
				jumpServerConf.SshUsername,
				jumpServerConf.SshPassword,
				jumpServerConf.SshKey,
				jumpServerConf.SshKeyPassphrase,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create jump server ssh client[%d]: %w", i+1, err)
			}
			defer newClient.Close()

			jumpClient = newClient
			d.logger.Info(fmt.Sprintf("jump server connected [%d]", i+1), slog.String("host", jumpServerConf.SshHost))
		}

		// 通过跳板机发起 TCP 连接到目标服务器
		targetConn, err = jumpClient.DialContext(ctx, "tcp", net.JoinHostPort(d.config.SshHost, strconv.Itoa(int(d.config.SshPort))))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to target server: %w", err)
		}
	} else {
		// 直接发起 TCP 连接到目标服务器
		targetConn, err = net.Dial("tcp", net.JoinHostPort(d.config.SshHost, strconv.Itoa(int(d.config.SshPort))))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to target server: %w", err)
		}
	}
	defer targetConn.Close()

	// 创建 SSH 客户端
	client, err := createSshClient(
		targetConn,
		d.config.SshHost,
		d.config.SshPort,
		d.config.SshAuthMethod,
		d.config.SshUsername,
		d.config.SshPassword,
		d.config.SshKey,
		d.config.SshKeyPassphrase,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh client: %w", err)
	}
	defer client.Close()
	d.logger.Info("ssh connected")

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

		stdout, stderr, err := execSshCommand(client, command)
		d.logger.Debug("run pre-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute pre-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	// 上传证书和私钥文件
	switch d.config.OutputFormat {
	case OUTPUT_FORMAT_PEM:
		if err := xssh.WriteRemoteString(client, d.config.OutputKeyPath, privkeyPEM, d.config.UseSCP); err != nil {
			return nil, fmt.Errorf("failed to upload private key file: %w", err)
		}
		d.logger.Info("ssl private key file uploaded", slog.String("path", d.config.OutputKeyPath))

		if err := xssh.WriteRemoteString(client, d.config.OutputCertPath, certPEM, d.config.UseSCP); err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		}
		d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))

		if d.config.OutputServerCertPath != "" {
			if err := xssh.WriteRemoteString(client, d.config.OutputServerCertPath, serverCertPEM, d.config.UseSCP); err != nil {
				return nil, fmt.Errorf("failed to save server certificate file: %w", err)
			}
			d.logger.Info("ssl server certificate file uploaded", slog.String("path", d.config.OutputServerCertPath))
		}

		if d.config.OutputIntermediaCertPath != "" {
			if err := xssh.WriteRemoteString(client, d.config.OutputIntermediaCertPath, intermediaCertPEM, d.config.UseSCP); err != nil {
				return nil, fmt.Errorf("failed to save intermedia certificate file: %w", err)
			}
			d.logger.Info("ssl intermedia certificate file uploaded", slog.String("path", d.config.OutputIntermediaCertPath))
		}

	case OUTPUT_FORMAT_PFX:
		pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
		}
		d.logger.Info("ssl certificate transformed to pfx")

		if err := xssh.WriteRemote(client, d.config.OutputCertPath, pfxData, d.config.UseSCP); err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		}
		d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))

	case OUTPUT_FORMAT_JKS:
		jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
		if err != nil {
			return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
		}
		d.logger.Info("ssl certificate transformed to jks")

		if err := xssh.WriteRemote(client, d.config.OutputCertPath, jksData, d.config.UseSCP); err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		}
		d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))

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

		stdout, stderr, err := execSshCommand(client, command)
		d.logger.Debug("run post-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute post-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	return &deployer.DeployResult{}, nil
}

func createSshClient(conn net.Conn, host string, port int32, authMethod string, username, password, key, keyPassphrase string) (*ssh.Client, error) {
	if host == "" {
		host = "localhost"
	}

	if port == 0 {
		port = 22
	}

	if username == "" {
		username = "root"
	}

	if authMethod == "" {
		if key != "" {
			authMethod = AUTH_METHOD_KEY
		} else if password != "" {
			authMethod = AUTH_METHOD_PASSWORD
		} else {
			authMethod = AUTH_METHOD_NONE
		}
	}

	switch authMethod {
	case AUTH_METHOD_NONE:
		return xssh.NewClient(conn, host, int(port), username)

	case AUTH_METHOD_PASSWORD:
		return xssh.NewClientWithPassword(conn, host, int(port), username, password)

	case AUTH_METHOD_KEY:
		return xssh.NewClientWithKey(conn, host, int(port), username, key, keyPassphrase)

	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethod)
	}
}

func execSshCommand(sshCli *ssh.Client, command string) (string, string, error) {
	session, err := sshCli.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	stdoutBuf := bytes.NewBuffer(nil)
	session.Stdout = stdoutBuf
	stderrBuf := bytes.NewBuffer(nil)
	session.Stderr = stderrBuf
	err = session.Run(command)
	if err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("failed to execute ssh command: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}
