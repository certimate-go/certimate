package ssh

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	conns []net.Conn
	clis  []*ssh.Client
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of SSH client is nil")
	}

	conns, clis, err := createConnsAndSshClients(config)
	if err != nil {
		for i := len(clis) - 1; i >= 0; i-- {
			clis[i].Close()
		}

		for i := len(conns) - 1; i >= 0; i-- {
			conns[i].Close()
		}

		return nil, fmt.Errorf("ssh: %w", err)
	}

	return &Client{conns: conns, clis: clis}, nil
}

func (c *Client) RawClient() *ssh.Client {
	if len(c.clis) == 0 {
		return nil
	}

	return c.clis[len(c.clis)-1]
}

func (c *Client) Close() error {
	errs := make([]error, 0)

	for i := len(c.clis) - 1; i >= 0; i-- {
		cli := c.clis[i]
		if err := cli.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	for i := len(c.conns) - 1; i >= 0; i-- {
		conn := c.conns[i]
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	} else if len(errs) == 1 {
		return fmt.Errorf("ssh: %w", errs[0])
	} else {
		return fmt.Errorf("ssh: %w", errors.Join(errs...))
	}
}

func createConnsAndSshClients(config *Config) (conns []net.Conn, clis []*ssh.Client, err error) {
	conns = make([]net.Conn, 0)
	clis = make([]*ssh.Client, 0)

	var targetConn net.Conn
	if len(config.JumpServers) > 0 {
		var jumpCli *ssh.Client

		for i, jumpConfig := range config.JumpServers {
			var jumpConn net.Conn
			if jumpCli == nil {
				jumpConn, err = net.Dial("tcp", resolveAddr(jumpConfig.Host, jumpConfig.Port))
			} else {
				jumpConn, err = jumpCli.Dial("tcp", resolveAddr(jumpConfig.Host, jumpConfig.Port))
			}
			if err != nil {
				err = fmt.Errorf("ssh: failed to connect to jump server [%d]: %w", i+1, err)
				return
			}

			conns = append(conns, jumpConn)

			jumpCli, err = createSshClientWithConn(&jumpConfig, jumpConn)
			if err != nil {
				err = fmt.Errorf("ssh: failed to create jump server SSH client[%d]: %w", i+1, err)
				return
			}

			clis = append(clis, jumpCli)
		}

		// 通过跳板机发起 TCP 连接到目标服务器
		targetConn, err = jumpCli.Dial("tcp", resolveAddr(config.Host, config.Port))
		if err != nil {
			err = fmt.Errorf("ssh: failed to connect to target server: %w", err)
			return
		}

		conns = append(conns, targetConn)
	} else {
		// 直接发起 TCP 连接到目标服务器
		targetConn, err = net.Dial("tcp", resolveAddr(config.Host, config.Port))
		if err != nil {
			err = fmt.Errorf("ssh: failed to connect to target server: %w", err)
			return
		}

		conns = append(conns, targetConn)
	}

	// 创建 SSH 客户端
	targetCli, err := createSshClientWithConn(&config.ServerConfig, targetConn)
	if err != nil {
		return nil, nil, fmt.Errorf("ssh: failed to create SSH client: %w", err)
	}

	clis = append(clis, targetCli)

	return conns, clis, nil
}

func createSshClientWithConn(config *ServerConfig, conn net.Conn) (*ssh.Client, error) {
	if conn == nil {
		return nil, fmt.Errorf("nil conn")
	}

	authMethodType := lo.
		If(string(config.AuthMethod) != "", config.AuthMethod).
		ElseIf(config.Key != "", AuthMethodTypeKey).
		ElseIf(config.Password != "", AuthMethodTypePassword).
		Else(AuthMethodTypeNone)
	authMethods := make([]ssh.AuthMethod, 0)
	switch authMethodType {
	case AuthMethodTypeNone:
		{
			if config.Username == "" {
				return nil, fmt.Errorf("missing username")
			}
		}

	case AuthMethodTypePassword:
		{
			if config.Username == "" {
				return nil, fmt.Errorf("missing username")
			}
			if config.Password == "" {
				return nil, fmt.Errorf("missing password")
			}

			password := config.Password
			authMethods = append(authMethods, ssh.Password(password))
			authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				if len(answers) == 0 {
					return answers, nil
				}

				for i, question := range questions {
					question = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(question), ":"))
					if strings.EqualFold(question, "Password") {
						answers[i] = password
						return answers, nil
					}
				}

				return nil, fmt.Errorf("unexpected keyboard interactive question '%s'", strings.Join(questions, ", "))
			}))
		}

	case AuthMethodTypeKey:
		{
			if config.Username == "" {
				return nil, fmt.Errorf("missing username")
			}
			if config.Key == "" {
				return nil, fmt.Errorf("missing key")
			}

			key := config.Key
			keyPassphrase := config.KeyPassphrase

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

			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}

	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", authMethodType)
	}

	addr := resolveAddr(config.Host, config.Port)
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

func resolveAddr(host string, port int) string {
	if port == 0 {
		port = defaultPort
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}
