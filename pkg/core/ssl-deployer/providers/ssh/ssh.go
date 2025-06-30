package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/sftp"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"

	"github.com/certimate-go/certimate/pkg/core"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type JumpServerConfig struct {
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

type SSLDeployerProviderConfig struct {
	// SSH 主机。
	// 零值时默认值 "localhost"。
	SshHost string `json:"sshHost,omitempty"`
	// SSH 端口。
	// 零值时默认值 22。
	SshPort int32 `json:"sshPort,omitempty"`
	// SSH 认证方式。
	// 可取值 "none"、"password" 或 "key"。
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
	// 跳板机配置数组。
	JumpServers []JumpServerConfig `json:"jumpServers,omitempty"`
	// 是否回退使用 SCP。
	UseSCP bool `json:"useSCP,omitempty"`
	// 前置命令。
	PreCommand string `json:"preCommand,omitempty"`
	// 后置命令。
	PostCommand string `json:"postCommand,omitempty"`
	// 输出证书格式。
	OutputFormat OutputFormatType `json:"outputFormat,omitempty"`
	// 输出证书文件路径。
	OutputCertPath string `json:"outputCertPath,omitempty"`
	// 输出服务器证书文件路径。
	// 选填。
	OutputServerCertPath string `json:"outputServerCertPath,omitempty"`
	// 输出中间证书文件路径。
	// 选填。
	OutputIntermediaCertPath string `json:"outputIntermediaCertPath,omitempty"`
	// 输出私钥文件路径。
	OutputKeyPath string `json:"outputKeyPath,omitempty"`
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

type SSLDeployerProvider struct {
	config *SSLDeployerProviderConfig
	logger *slog.Logger
}

var _ core.SSLDeployer = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	return &SSLDeployerProvider{
		config: config,
		logger: slog.Default(),
	}, nil
}

func (d *SSLDeployerProvider) SetLogger(logger *slog.Logger) {
	if logger == nil {
		d.logger = slog.New(slog.DiscardHandler)
	} else {
		d.logger = logger
	}
}

func (d *SSLDeployerProvider) Deploy(ctx context.Context, certPEM string, privkeyPEM string) (*core.SSLDeployResult, error) {
	// 提取服务器证书和中间证书
	serverCertPEM, intermediaCertPEM, err := xcert.ExtractCertificatesFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract certs: %w", err)
	}

	var targetConn net.Conn

	// 连接到跳板机
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

	// 通过已有的连接创建目标服务器 SSH 客户端
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
		stdout, stderr, err := execSshCommand(client, d.config.PreCommand)
		d.logger.Debug("run pre-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute pre-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	// 上传证书和私钥文件
	switch d.config.OutputFormat {
	case OUTPUT_FORMAT_PEM:
		if err := writeFileString(client, d.config.UseSCP, d.config.OutputCertPath, certPEM); err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		}
		d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))

		if d.config.OutputServerCertPath != "" {
			if err := writeFileString(client, d.config.UseSCP, d.config.OutputServerCertPath, serverCertPEM); err != nil {
				return nil, fmt.Errorf("failed to save server certificate file: %w", err)
			}
			d.logger.Info("ssl server certificate file uploaded", slog.String("path", d.config.OutputServerCertPath))
		}

		if d.config.OutputIntermediaCertPath != "" {
			if err := writeFileString(client, d.config.UseSCP, d.config.OutputIntermediaCertPath, intermediaCertPEM); err != nil {
				return nil, fmt.Errorf("failed to save intermedia certificate file: %w", err)
			}
			d.logger.Info("ssl intermedia certificate file uploaded", slog.String("path", d.config.OutputIntermediaCertPath))
		}

		if err := writeFileString(client, d.config.UseSCP, d.config.OutputKeyPath, privkeyPEM); err != nil {
			return nil, fmt.Errorf("failed to upload private key file: %w", err)
		}
		d.logger.Info("ssl private key file uploaded", slog.String("path", d.config.OutputKeyPath))

	case OUTPUT_FORMAT_PFX:
		pfxData, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, d.config.PfxPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to transform certificate to PFX: %w", err)
		}
		d.logger.Info("ssl certificate transformed to pfx")

		if err := writeFile(client, d.config.UseSCP, d.config.OutputCertPath, pfxData); err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		}
		d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))

	case OUTPUT_FORMAT_JKS:
		jksData, err := xcert.TransformCertificateFromPEMToJKS(certPEM, privkeyPEM, d.config.JksAlias, d.config.JksKeypass, d.config.JksStorepass)
		if err != nil {
			return nil, fmt.Errorf("failed to transform certificate to JKS: %w", err)
		}
		d.logger.Info("ssl certificate transformed to jks")

		if err := writeFile(client, d.config.UseSCP, d.config.OutputCertPath, jksData); err != nil {
			return nil, fmt.Errorf("failed to upload certificate file: %w", err)
		}
		d.logger.Info("ssl certificate file uploaded", slog.String("path", d.config.OutputCertPath))

	default:
		return nil, fmt.Errorf("unsupported output format '%s'", d.config.OutputFormat)
	}

	// 执行后置命令
	if d.config.PostCommand != "" {
		stdout, stderr, err := execSshCommand(client, d.config.PostCommand)
		d.logger.Debug("run post-command", slog.String("stdout", stdout), slog.String("stderr", stderr))
		if err != nil {
			return nil, fmt.Errorf("failed to execute post-command (stdout: %s, stderr: %s): %w ", stdout, stderr, err)
		}
	}

	return &core.SSLDeployResult{}, nil
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

	const AUTH_METHOD_NONE = "none"
	const AUTH_METHOD_PASSWORD = "password"
	const AUTH_METHOD_KEY = "key"
	if authMethod == "" {
		if key != "" {
			authMethod = AUTH_METHOD_KEY
		} else if password != "" {
			authMethod = AUTH_METHOD_PASSWORD
		} else {
			authMethod = AUTH_METHOD_NONE
		}
	}

	authentications := make([]ssh.AuthMethod, 0)
	switch authMethod {
	case AUTH_METHOD_NONE:
		{
		}

	case AUTH_METHOD_PASSWORD:
		{
			authentications = append(authentications, ssh.Password(password))
			authentications = append(authentications, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				if len(questions) == 1 {
					return []string{password}, nil
				}
				return nil, fmt.Errorf("unexpected keyboard interactive question [%s]", strings.Join(questions, ", "))
			}))
		}

	case AUTH_METHOD_KEY:
		{
			var signer ssh.Signer
			var err error

			if keyPassphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(keyPassphrase))
			} else {
				signer, err = ssh.ParsePrivateKey([]byte(key))
			}

			if err != nil {
				return nil, err
			}

			authentications = append(authentications, ssh.PublicKeys(signer))
		}

	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethod)
	}

	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, &ssh.ClientConfig{
		User:            username,
		Auth:            authentications,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
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

func writeFileString(sshCli *ssh.Client, useSCP bool, path string, content string) error {
	if useSCP {
		return writeFileStringWithSCP(sshCli, path, content)
	}

	return writeFileStringWithSFTP(sshCli, path, content)
}

func writeFile(sshCli *ssh.Client, useSCP bool, path string, data []byte) error {
	if useSCP {
		return writeFileWithSCP(sshCli, path, data)
	}

	return writeFileWithSFTP(sshCli, path, data)
}

func writeFileStringWithSCP(sshCli *ssh.Client, path string, content string) error {
	return writeFileWithSCP(sshCli, path, []byte(content))
}

func writeFileWithSCP(sshCli *ssh.Client, path string, data []byte) error {
	scpCli, err := scp.NewClientFromExistingSSH(sshCli, &scp.ClientOption{})
	if err != nil {
		return fmt.Errorf("failed to create scp client: %w", err)
	}

	reader := bytes.NewReader(data)
	err = scpCli.CopyToRemote(reader, path, &scp.FileTransferOption{})
	if err != nil {
		return fmt.Errorf("failed to write to remote file: %w", err)
	}

	return nil
}

func writeFileStringWithSFTP(sshCli *ssh.Client, path string, content string) error {
	return writeFileWithSFTP(sshCli, path, []byte(content))
}

func writeFileWithSFTP(sshCli *ssh.Client, path string, data []byte) error {
	sftpCli, err := sftp.NewClient(sshCli)
	if err != nil {
		return fmt.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpCli.Close()

	if err := sftpCli.MkdirAll(filepath.ToSlash(filepath.Dir(path))); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	file, err := sftpCli.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to remote file: %w", err)
	}

	return nil
}
