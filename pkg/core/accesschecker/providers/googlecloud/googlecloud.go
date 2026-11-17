package googlecloud

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/accesschecker"
)

type (
	Provider    = core.AccessChecker
	CheckResult = core.AccessCheckResult
)

const (
	defaultTokenURL = "https://oauth2.googleapis.com/token"
	defaultDNSURL   = "https://dns.googleapis.com"
	tokenAudience   = "https://oauth2.googleapis.com/token"
	dnsScope        = "https://www.googleapis.com/auth/ndev.clouddns.readonly"
)

type checker struct {
	tokenURL string
	dnsURL   string
}

var _ Provider = (*checker)(nil)

func New() Provider {
	return &checker{tokenURL: defaultTokenURL, dnsURL: defaultDNSURL}
}

func (c *checker) ProviderType() string {
	return string(domain.AccessProviderTypeGoogleCloud)
}

func (c *checker) Scope() []string {
	return []string{core.AccessCheckItemConnect, core.AccessCheckItemAuth}
}

func (c *checker) Check(ctx context.Context, config map[string]any) (*CheckResult, error) {
	ctx, cancel := accesschecker.WithBudget(ctx)
	defer cancel()

	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	tokenURL := c.tokenURL
	dnsURL := c.dnsURL
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	if dnsURL == "" {
		dnsURL = defaultDNSURL
	}

	diagnostics := []string{accesschecker.HostOf(tokenURL)}
	if accesschecker.HostOf(dnsURL) != diagnostics[0] {
		diagnostics = append(diagnostics, accesschecker.HostOf(dnsURL))
	}
	secrets := []string{cfg.PrivateKey, cfg.ClientEmail}

	assertion, err := signJWT(cfg.ClientEmail, cfg.PrivateKey, time.Now())
	if err != nil {
		return nil, fmt.Errorf("%w: serviceAccountKey is invalid", core.ErrAccessCheckConfigInvalid)
	}
	secrets = append(secrets, assertion)

	token, status, code, callErr := exchangeToken(ctx, tokenURL, assertion)
	if callErr != nil || token == "" {
		return accesschecker.Finish(callErr, status, code, diagnostics, secrets...), nil
	}
	secrets = append(secrets, token)

	status, callErr = listManagedZones(ctx, dnsURL, cfg.ProjectId, token)
	if callErr != nil {
		return accesschecker.Finish(callErr, status, "", diagnostics, secrets...), nil
	}
	if status == http.StatusForbidden {
		diag := append(append([]string{}, diagnostics...), fmt.Sprintf("HTTP %d", status))
		return accesschecker.Redact(accesschecker.Passed(diag), secrets...), nil
	}
	return accesschecker.Finish(nil, status, "", diagnostics, secrets...), nil
}

type checkConfig struct {
	ProjectId   string
	ClientEmail string
	PrivateKey  string
}

type serviceAccountKey struct {
	Type        string `json:"type"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

func parseConfig(config map[string]any) (*checkConfig, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("%w: config is empty", core.ErrAccessCheckConfigInvalid)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w: config is not serializable", core.ErrAccessCheckConfigInvalid)
	}

	var wire domain.AccessConfigForGoogleCloud
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("%w: config fields are malformed", core.ErrAccessCheckConfigInvalid)
	}

	projectId := strings.TrimSpace(wire.ProjectId)
	if projectId == "" || !safeProject(projectId) {
		return nil, fmt.Errorf("%w: projectId is required", core.ErrAccessCheckConfigInvalid)
	}

	var key serviceAccountKey
	if err := json.Unmarshal([]byte(wire.ServiceAccountKey), &key); err != nil {
		return nil, fmt.Errorf("%w: serviceAccountKey is invalid", core.ErrAccessCheckConfigInvalid)
	}
	if key.Type != "" && key.Type != "service_account" {
		return nil, fmt.Errorf("%w: serviceAccountKey is invalid", core.ErrAccessCheckConfigInvalid)
	}
	email := strings.TrimSpace(key.ClientEmail)
	if email == "" || strings.TrimSpace(key.PrivateKey) == "" {
		return nil, fmt.Errorf("%w: serviceAccountKey is invalid", core.ErrAccessCheckConfigInvalid)
	}
	if _, err := parsePrivateKey(key.PrivateKey); err != nil {
		return nil, fmt.Errorf("%w: serviceAccountKey is invalid", core.ErrAccessCheckConfigInvalid)
	}
	return &checkConfig{ProjectId: projectId, ClientEmail: email, PrivateKey: key.PrivateKey}, nil
}

func signJWT(email, privateKeyPEM string, now time.Time) (string, error) {
	key, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(map[string]any{
		"iss":   email,
		"scope": dnsScope,
		"aud":   tokenAudience,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	signingInput := header + "." + payload
	sum := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func parsePrivateKey(pemText string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, fmt.Errorf("pem")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err1 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err1 != nil {
			return nil, err
		}
		return key, nil
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not rsa")
	}
	return key, nil
}

func exchangeToken(ctx context.Context, tokenURL, assertion string) (string, int, string, error) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, "", fmt.Errorf("%w: token url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", resp.StatusCode, "", err
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && payload.AccessToken != "" {
		return payload.AccessToken, resp.StatusCode, "", nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "", 0, "", fmt.Errorf("token endpoint returned no credential verdict")
	}
	return "", resp.StatusCode, payload.Error, fmt.Errorf("token endpoint rejected the credential")
}

func listManagedZones(ctx context.Context, dnsURL, projectId, token string) (int, error) {
	target := strings.TrimRight(dnsURL, "/") + "/dns/v1/projects/" + url.PathEscape(projectId) + "/managedZones?maxResults=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: dns url is invalid", core.ErrAccessCheckConfigInvalid)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := accesschecker.HTTPClient().Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, nil
}

func safeProject(value string) bool {
	if strings.Contains(value, "..") {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.' || r == ':':
		default:
			return false
		}
	}
	return true
}
