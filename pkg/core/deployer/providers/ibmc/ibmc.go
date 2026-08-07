package ibmc

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/certimate-go/certimate/pkg/core"
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
	client *http.Client
}

var _ Provider = (*Deployer)(nil)

type sessionResponse struct {
	Token string `json:"X-Auth-Token"`
}

type redfishError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

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
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: config.AllowInsecureConnections} //nolint:gosec // explicitly configured for iBMC self-signed certificates
	return &Deployer{
		config: config,
		logger: slog.Default(),
		client: &http.Client{Transport: transport, Timeout: 30 * time.Second},
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
	payload := map[string]string{"Certificate": base64.StdEncoding.EncodeToString(pfx), "Password": certificatePassword}
	var failures []string
	for _, endpoint := range strings.Split(d.config.Endpoint, "\n") {
		if err := d.deployHost(ctx, strings.TrimSpace(endpoint), payload); err != nil {
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

func (d *Deployer) deployHost(ctx context.Context, base string, payload map[string]string) error {
	token, sessionURL, err := d.createSession(ctx, base)
	if err != nil {
		return fmt.Errorf("failed to create Redfish session: %w", err)
	}
	defer func() { _ = d.deleteSession(context.Background(), token, sessionURL) }()
	managerPath, err := d.discoverManager(ctx, base, token)
	if err != nil {
		return err
	}
	path := managerPath + "/SecurityService/HttpsCert/Actions/HttpsCert.ImportCustomCertificate"
	if _, err := d.doJSON(ctx, http.MethodPost, base+path, token, payload, nil); err != nil {
		return fmt.Errorf("failed to import HTTPS certificate: %w", err)
	}
	d.logger.Info("iBMC HTTPS certificate imported", slog.String("endpoint", base), slog.String("manager", managerPath))
	if d.config.RestartAfterImport {
		if _, err := d.doJSON(ctx, http.MethodPost, base+managerPath+"/Actions/Manager.Reset", token, map[string]string{"ResetType": "ForceRestart"}, nil); err != nil {
			return fmt.Errorf("failed to restart iBMC after certificate import: %w", err)
		}
		d.logger.Info("iBMC restart requested", slog.String("endpoint", base), slog.String("manager", managerPath))
	}
	return nil
}

func (d *Deployer) discoverManager(ctx context.Context, base, token string) (string, error) {
	response := &managersResponse{}
	if _, err := d.doJSON(ctx, http.MethodGet, base+"/redfish/v1/Managers", token, nil, response); err != nil {
		return "", fmt.Errorf("failed to query Redfish managers: %w", err)
	}
	if len(response.Members) == 0 {
		return "", fmt.Errorf("Redfish managers collection is empty")
	}
	manager := response.Members[0].ODataID
	if manager == "" && response.Members[0].ID != "" {
		manager = "/redfish/v1/Managers/" + url.PathEscape(response.Members[0].ID)
	}
	if manager == "" {
		return "", fmt.Errorf("Redfish manager has no resource ID")
	}
	if managerURL, err := url.Parse(manager); err == nil && managerURL.IsAbs() {
		manager = managerURL.EscapedPath()
		if managerURL.RawQuery != "" {
			manager += "?" + managerURL.RawQuery
		}
	} else if !strings.HasPrefix(manager, "/") {
		manager = "/" + manager
	}
	return strings.TrimRight(manager, "/"), nil
}

func (d *Deployer) createSession(ctx context.Context, base string) (string, string, error) {
	response := &sessionResponse{}
	resp, err := d.doJSON(ctx, http.MethodPost, base+"/redfish/v1/SessionService/Sessions", "", map[string]string{"UserName": d.config.Username, "Password": d.config.Password}, response)
	if err != nil {
		return "", "", err
	}
	token := resp.Header.Get("X-Auth-Token")
	if token == "" {
		token = response.Token
	}
	if token == "" {
		return "", "", fmt.Errorf("iBMC did not return X-Auth-Token")
	}
	location := resp.Header.Get("Location")
	if location != "" {
		u, err := url.Parse(location)
		if err == nil && !u.IsAbs() {
			if baseURL, err := url.Parse(base); err == nil {
				location = baseURL.ResolveReference(u).String()
			}
		}
	}
	return token, location, nil
}

func (d *Deployer) deleteSession(ctx context.Context, token, location string) error {
	if location == "" {
		return nil
	}
	_, err := d.doJSON(ctx, http.MethodDelete, location, token, nil, nil)
	return err
}

func (d *Deployer) doJSON(ctx context.Context, method, endpoint, token string, payload, result any) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("X-Auth-Token", token)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr redfishError
		if json.Unmarshal(data, &apiErr) == nil && (apiErr.Error.Code != "" || apiErr.Error.Message != "") {
			return nil, fmt.Errorf("HTTP %d: %s: %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if result != nil && len(data) > 0 {
		if err := json.Unmarshal(data, result); err != nil {
			return nil, fmt.Errorf("invalid iBMC response: %w", err)
		}
	}
	return resp, nil
}
