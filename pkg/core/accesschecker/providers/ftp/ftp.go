package ftp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	ftplib "github.com/jlaffaye/ftp"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
	CheckItem   = core.AccessCheckItem
)

const defaultPort = 21

type checker struct{}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeFTP)
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	diagnostics := []string{cfg.addr()}
	conn, err := dialControl(ctx, cfg.addr())
	if err != nil {
		if result := contextResult(ctx, diagnostics); result != nil {
			return result, nil
		}
		return connectFailedResult(diagnostics), nil
	}
	defer conn.Quit()

	if err := conn.Login(cfg.Username, cfg.Password); err != nil {
		if result := contextResult(ctx, diagnostics); result != nil {
			return result, nil
		}
		return authRejectedResult(append(diagnostics, "login rejected")), nil
	}

	return passedResult(append(diagnostics, "credential accepted; permissions on specific resources were not checked")), nil
}

type checkConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForFTP
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	cfg := &checkConfig{
		Host:     strings.TrimSpace(wire.Host),
		Port:     int(wire.Port),
		Username: strings.TrimSpace(wire.Username),
		Password: wire.Password,
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("%w: host is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("%w: port is out of range", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("%w: username is required", core.ErrAccessCheckConfigInvalid)
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("%w: password is required", core.ErrAccessCheckConfigInvalid)
	}
	if strings.ContainsAny(cfg.Username, "\r\n") || strings.ContainsAny(cfg.Password, "\r\n") {
		return nil, fmt.Errorf("%w: username or password contains a line break", core.ErrAccessCheckConfigInvalid)
	}

	return cfg, nil
}

func (c *checkConfig) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func dialControl(ctx context.Context, addr string) (*ftplib.ServerConn, error) {
	return ftplib.Dial(addr, ftplib.DialWithDialFunc(func(network, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		if deadline, ok := ctx.Deadline(); ok {
			_ = conn.SetDeadline(deadline)
		}
		go func() {
			<-ctx.Done()
			_ = conn.Close()
		}()
		return conn, nil
	}))
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

func authRejectedResult(diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusFailed,
		ReasonCode: core.AccessCheckReasonAuthRejected,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusPassed},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusFailed},
		},
		Diagnostics: diagnostics,
	}
}

func inconclusiveResult(reason string, diagnostics []string) *CheckResult {
	return &CheckResult{
		Status:     core.AccessCheckStatusInconclusive,
		ReasonCode: reason,
		Items: []CheckItem{
			{Id: core.AccessCheckItemConnect, Status: core.AccessCheckItemStatusSkipped},
			{Id: core.AccessCheckItemAuth, Status: core.AccessCheckItemStatusSkipped},
		},
		Diagnostics: diagnostics,
	}
}

func contextResult(ctx context.Context, diagnostics []string) *CheckResult {
	err := ctx.Err()
	if err == nil {
		return nil
	}
	reason := core.AccessCheckReasonTimeout
	if errors.Is(err, context.Canceled) {
		reason = core.AccessCheckReasonCanceled
	}
	return inconclusiveResult(reason, diagnostics)
}
