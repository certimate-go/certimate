package ssh

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/crypto/ssh"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
	CheckItem   = core.AccessCheckItem
)

const (
	sshAuthMethodNone     = "none"
	sshAuthMethodPassword = "password"
	sshAuthMethodKey      = "key"

	sshDefaultPort   = 22
	sshClientVersion = "SSH-2.0-Certimate_AccessCheck_1.0"
)

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeSSH)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, core.AccessCheckDefaultBudget)
		defer cancel()
	}

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	chain := append(slices.Clone(cfg.JumpServers), cfg.Target)
	auths := make([][]ssh.AuthMethod, len(chain))
	badCredentialHop := -1
	for i := range chain {
		auths[i], err = buildAuthMethods(&chain[i])
		if err != nil {
			badCredentialHop = i
			break
		}
	}
	if badCredentialHop >= 0 {
		entry := &chain[0]
		diagnostics := []string{entry.addr(), "authMethod=" + entry.AuthMethod}
		if badCredentialHop == len(chain)-1 {
			diagnostics = append(diagnostics, "invalid credential material at target")
		} else {
			diagnostics = append(diagnostics, fmt.Sprintf("invalid credential material at jump server [%d]", badCredentialHop+1))
		}

		_, probeErr := probeEndpoint(ctx, entry)
		if ctx.Err() != nil {
			return inconclusiveResult(ctx.Err(), diagnostics), nil
		}
		if probeErr != nil {
			return connectFailedResult(diagnostics), nil
		}
		return authRejectedResult(diagnostics, true), nil
	}

	connChain, failedHop, dialErr := dialChain(ctx, chain, auths)
	if connChain != nil {
		defer connChain.Close()
	}

	diagnostics := []string{cfg.Target.addr(), "authMethod=" + cfg.Target.AuthMethod}
	if ctx.Err() != nil {
		return inconclusiveResult(ctx.Err(), diagnostics), nil
	}
	if dialErr != nil {
		if isAuthFailure(dialErr) {
			return authRejectedResult(diagnostics, failedHop == len(chain)-1), nil
		}
		return connectFailedResult(diagnostics), nil
	}

	return passedResult(diagnostics), nil
}

type configWire struct {
	Host           string       `json:"host"`
	Port           int32        `json:"port"`
	AuthMethod     string       `json:"authMethod"`
	Username       string       `json:"username"`
	Password       string       `json:"password,omitempty"`
	Key            string       `json:"key,omitempty"`
	KeyPassphrase  string       `json:"keyPassphrase,omitempty"`
	KeyCertificate string       `json:"keyCertificate,omitempty"`
	JumpServers    []configWire `json:"jumpServers,omitempty"`
}

type checkConfig struct {
	Target      endpointConfig
	JumpServers []endpointConfig
}

type endpointConfig struct {
	Host           string
	Port           int
	AuthMethod     string
	Username       string
	Password       string
	Key            string
	KeyPassphrase  string
	KeyCertificate string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire configWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	target, err := wire.toEndpointConfig()
	if err != nil {
		return nil, err
	}

	cfg := &checkConfig{Target: *target}
	for i := range wire.JumpServers {
		jump, err := wire.JumpServers[i].toEndpointConfig()
		if err != nil {
			return nil, fmt.Errorf("%w: jump server [%d]: %s", core.ErrAccessCheckConfigInvalid, i+1, err.Error())
		}
		cfg.JumpServers = append(cfg.JumpServers, *jump)
	}

	return cfg, nil
}

func (w *configWire) toEndpointConfig() (*endpointConfig, error) {
	e := &endpointConfig{
		Host:           strings.TrimSpace(w.Host),
		Port:           int(w.Port),
		AuthMethod:     w.AuthMethod,
		Username:       w.Username,
		Password:       w.Password,
		Key:            w.Key,
		KeyPassphrase:  w.KeyPassphrase,
		KeyCertificate: w.KeyCertificate,
	}

	if e.Host == "" {
		return nil, fmt.Errorf("%w: host is required", core.ErrAccessCheckConfigInvalid)
	}
	if e.Port == 0 {
		e.Port = sshDefaultPort
	}
	if e.Port < 1 || e.Port > 65535 {
		return nil, fmt.Errorf("%w: port is out of range", core.ErrAccessCheckConfigInvalid)
	}

	if e.AuthMethod == "" {
		e.AuthMethod = lo.
			If(e.Key != "", sshAuthMethodKey).
			ElseIf(e.Password != "", sshAuthMethodPassword).
			Else(sshAuthMethodNone)
	}
	switch e.AuthMethod {
	case sshAuthMethodNone, sshAuthMethodPassword, sshAuthMethodKey:
	default:
		return nil, fmt.Errorf("%w: authMethod is invalid", core.ErrAccessCheckConfigInvalid)
	}

	if e.Username == "" {
		return nil, fmt.Errorf("%w: username is required", core.ErrAccessCheckConfigInvalid)
	}
	switch e.AuthMethod {
	case sshAuthMethodPassword:
		if e.Password == "" {
			return nil, fmt.Errorf("%w: password is required", core.ErrAccessCheckConfigInvalid)
		}
	case sshAuthMethodKey:
		if e.Key == "" {
			return nil, fmt.Errorf("%w: key is required", core.ErrAccessCheckConfigInvalid)
		}
	}

	return e, nil
}

func (e *endpointConfig) addr() string {
	port := e.Port
	if port == 0 {
		port = sshDefaultPort
	}
	return net.JoinHostPort(e.Host, strconv.Itoa(port))
}

type connChain struct {
	conns []net.Conn
	clis  []*ssh.Client
}

func (c *connChain) Close() error {
	errs := make([]error, 0, len(c.conns)+len(c.clis))
	for i := len(c.clis) - 1; i >= 0; i-- {
		if err := c.clis[i].Close(); err != nil {
			errs = append(errs, err)
		}
	}
	for i := len(c.conns) - 1; i >= 0; i-- {
		if err := c.conns[i].Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func dialChain(ctx context.Context, chain []endpointConfig, auths [][]ssh.AuthMethod) (*connChain, int, error) {
	cc := &connChain{}

	watcherDone := make(chan struct{})
	defer close(watcherDone)
	go func() {
		select {
		case <-ctx.Done():
			cc.Close()
		case <-watcherDone:
		}
	}()

	var prevCli *ssh.Client
	for i := range chain {
		ep := &chain[i]

		var conn net.Conn
		var err error
		if prevCli == nil {
			dialer := net.Dialer{}
			conn, err = dialer.DialContext(ctx, "tcp", ep.addr())
		} else {
			conn, err = prevCli.DialContext(ctx, "tcp", ep.addr())
		}
		if err != nil {
			cc.Close()
			return nil, i, err
		}
		cc.conns = append(cc.conns, conn)

		if deadline, ok := ctx.Deadline(); ok {
			conn.SetDeadline(deadline)
		}

		cli, err := newClientOnConn(ep, conn, auths[i])
		if err != nil {
			conn.Close()
			cc.Close()
			return nil, i, err
		}
		cc.clis = append(cc.clis, cli)
		prevCli = cli
	}

	return cc, -1, nil
}

func newClientOnConn(ep *endpointConfig, conn net.Conn, authMethods []ssh.AuthMethod) (*ssh.Client, error) {
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, ep.addr(), &ssh.ClientConfig{
		User:            ep.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

func probeEndpoint(ctx context.Context, ep *endpointConfig) (bool, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", ep.addr())
	if err != nil {
		return false, err
	}
	defer conn.Close()

	watcherDone := make(chan struct{})
	defer close(watcherDone)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-watcherDone:
		}
	}()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	if _, err := conn.Write([]byte(sshClientVersion + "\r\n")); err != nil {
		return false, err
	}

	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(line, "SSH-") {
		return false, fmt.Errorf("endpoint did not present an ssh banner")
	}

	return true, nil
}

func buildAuthMethods(ep *endpointConfig) ([]ssh.AuthMethod, error) {
	switch ep.AuthMethod {
	case sshAuthMethodNone:
		return nil, nil

	case sshAuthMethodPassword:
		return []ssh.AuthMethod{
			ssh.Password(ep.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				if len(answers) == 0 {
					return answers, nil
				}

				password := ep.Password
				for i, question := range questions {
					question = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(question), ":"))
					if strings.EqualFold(question, "Password") {
						answers[i] = password
						return answers, nil
					}
				}

				return nil, fmt.Errorf("unexpected keyboard interactive question '%s'", strings.Join(questions, ", "))
			}),
		}, nil

	case sshAuthMethodKey:
		signers, err := createSigners(ep.Key, ep.KeyPassphrase, ep.KeyCertificate)
		if err != nil {
			return nil, err
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signers...)}, nil

	default:
		return nil, fmt.Errorf("unsupported auth method '%s'", ep.AuthMethod)
	}
}

func createSigners(key, keyPassphrase, keyCertificate string) ([]ssh.Signer, error) {
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

	keyCertificate = strings.TrimSpace(keyCertificate)
	if keyCertificate == "" {
		return []ssh.Signer{signer}, nil
	}

	parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyCertificate))
	if err != nil {
		return nil, err
	}

	cert, ok := parsed.(*ssh.Certificate)
	if !ok {
		return nil, fmt.Errorf("key certificate is not an open-ssh certificate")
	}
	if cert.CertType != ssh.UserCert {
		return nil, fmt.Errorf("key certificate is not a user certificate")
	}

	certSigner, err := ssh.NewCertSigner(cert, signer)
	if err != nil {
		return nil, err
	}

	return []ssh.Signer{certSigner, signer}, nil
}

func isAuthFailure(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unable to authenticate")
}

func passedResult(diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusPassed,
		ReasonCode: "",
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusPassed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusPassed},
		},
		Diagnostics: diagnostics,
	}
}

func connectFailedResult(diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: core.AccessCheckReasonConnectFailed,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusFailed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}

func authRejectedResult(diagnostics []string, connectPassed bool) *CheckResult {
	connectStatus := lo.Ternary(connectPassed, core.AccessCheckItemStatusPassed, core.AccessCheckItemStatusSkipped)
	return &CheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: core.AccessCheckReasonAuthRejected,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: connectStatus},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusFailed},
		},
		Diagnostics: diagnostics,
	}
}

func inconclusiveResult(ctxErr error, diagnostics []string) *CheckResult {
	reasonCode := lo.Ternary(errors.Is(ctxErr, context.Canceled), core.AccessCheckReasonCanceled, core.AccessCheckReasonTimeout)
	return &CheckResult{
		Status:     core.AccessCheckStatusInconclusive,
		ReasonCode: reasonCode,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusSkipped},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}
