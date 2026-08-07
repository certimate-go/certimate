package ibmc

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"

	"github.com/certimate-go/certimate/pkg/core"
	ibmcsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ibmc"
	xcert "github.com/certimate-go/certimate/pkg/utils/cert"
)

type Provider = core.Deployer
type DeployResult = core.DeployerDeployResult

type DeployerConfig struct {
	Endpoint                 string `json:"endpoint"`
	Username                 string `json:"username"`
	Password                 string `json:"password"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections"`
	RestartAfterImport       bool   `json:"restartAfterImport"`
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
	if config.Endpoint == "" || config.Username == "" || config.Password == "" {
		return nil, fmt.Errorf("iBMC endpoint, username, and password are required")
	}
	endpoints, err := normalizeEndpoints(config.Endpoint)
	if err != nil {
		return nil, err
	}
	config.Endpoint = strings.Join(endpoints, "\n")
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
	certificatePassword, err := generateCertificatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCS#12 password: %w", err)
	}
	pfx, err := xcert.TransformCertificateFromPEMToPFX(certPEM, privkeyPEM, certificatePassword, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert certificate to PKCS#12: %w", err)
	}
	var failures []string
	for _, endpoint := range strings.Split(d.config.Endpoint, "\n") {
		if err := d.deployHost(ctx, strings.TrimSpace(endpoint), pfx, certificatePassword); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", endpoint, err))
		}
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("iBMC deployment failed for %d host(s): %s", len(failures), strings.Join(failures, "; "))
	}
	return &DeployResult{}, nil
}

func generateCertificatePassword() (string, error) {
	password := make([]byte, 24)
	if _, err := crand.Read(password); err != nil {
		return "", err
	}
	return hex.EncodeToString(password), nil
}

func normalizeEndpoints(raw string) ([]string, error) {
	var endpoints []string
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		base, err := normalizeEndpoint(line)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, base)
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("at least one iBMC endpoint is required")
	}
	return endpoints, nil
}

func normalizeEndpoint(raw string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(raw), "/")
	if !strings.Contains(base, "://") {
		if net.ParseIP(base) != nil && strings.Contains(base, ":") {
			base = "[" + base + "]"
		}
		base = "https://" + base
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return "", fmt.Errorf("iBMC endpoint must be a host or http(s) URL")
	}
	return base, nil
}

type managersResponse struct {
	Members []struct {
		ID      string `json:"Id"`
		ODataID string `json:"@odata.id"`
	} `json:"Members"`
}

func (d *Deployer) deployHost(ctx context.Context, base string, certificate []byte, password string) error {
	client := ibmcsdk.NewClient(base, d.config.Username, d.config.Password, d.config.AllowInsecureConnections)
	token, sessionURL, err := client.CreateSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Redfish session: %w", err)
	}
	defer func() { _ = client.DeleteSession(context.Background(), token, sessionURL) }()
	managerPath, err := client.DiscoverManager(ctx, token)
	if err != nil {
		return err
	}
	if err := client.ImportCustomCertificate(ctx, token, managerPath, certificate, password); err != nil {
		return fmt.Errorf("failed to import HTTPS certificate: %w", err)
	}
	d.logger.Info("iBMC HTTPS certificate imported", slog.String("endpoint", base), slog.String("manager", managerPath))
	if d.config.RestartAfterImport {
		if err := client.ResetManager(ctx, token, managerPath); err != nil {
			return fmt.Errorf("failed to restart iBMC after certificate import: %w", err)
		}
		d.logger.Info("iBMC restart requested", slog.String("endpoint", base), slog.String("manager", managerPath))
	}
	return nil
}
