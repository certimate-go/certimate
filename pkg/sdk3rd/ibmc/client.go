package ibmc

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const redfishDocumentationURL = "https://www.dmtf.org/standards/redfish"

type Client struct {
	rc       *resty.Client
	username string
	password string
}

type sessionResponse struct {
	Token string `json:"X-Auth-Token"`
}

type managersResponse struct {
	Members []struct {
		ID      string `json:"Id"`
		ODataID string `json:"@odata.id"`
	} `json:"Members"`
}

type redfishError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a Redfish client for one iBMC host.
func NewClient(endpoint, username, password string, allowInsecureConnections bool) *Client {
	transport := resty.New().
		SetBaseURL(strings.TrimRight(endpoint, "/")).
		SetTimeout(30 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: allowInsecureConnections}) //nolint:gosec // explicitly configured for iBMC self-signed certificates
	return &Client{rc: transport, username: username, password: password}
}

// CreateSession calls Redfish SessionService/Sessions.
// Documentation: https://www.dmtf.org/standards/redfish
func (c *Client) CreateSession(ctx context.Context) (token, location string, err error) {
	result := &sessionResponse{}
	response, err := c.do(ctx, resty.MethodPost, "/redfish/v1/SessionService/Sessions", "", map[string]string{
		"UserName": c.username,
		"Password": c.password,
	}, result)
	if err != nil {
		return "", "", err
	}
	token = response.Header().Get("X-Auth-Token")
	if token == "" {
		token = result.Token
	}
	if token == "" {
		return "", "", fmt.Errorf("iBMC did not return X-Auth-Token")
	}
	location = response.Header().Get("Location")
	if location != "" {
		location = c.absoluteURL(location)
	}
	return token, location, nil
}

// DeleteSession calls Redfish SessionService/Sessions/{id}.
// Documentation: https://www.dmtf.org/standards/redfish
func (c *Client) DeleteSession(ctx context.Context, token, location string) error {
	if location == "" {
		return nil
	}
	_, err := c.do(ctx, resty.MethodDelete, location, token, nil, nil)
	return err
}

// DiscoverManager calls Redfish Managers and returns the first manager resource path.
// Documentation: https://www.dmtf.org/standards/redfish
func (c *Client) DiscoverManager(ctx context.Context, token string) (string, error) {
	result := &managersResponse{}
	if _, err := c.do(ctx, resty.MethodGet, "/redfish/v1/Managers", token, nil, result); err != nil {
		return "", fmt.Errorf("failed to query Redfish managers: %w", err)
	}
	if len(result.Members) == 0 {
		return "", fmt.Errorf("Redfish managers collection is empty")
	}
	manager := result.Members[0].ODataID
	if manager == "" && result.Members[0].ID != "" {
		manager = "/redfish/v1/Managers/" + url.PathEscape(result.Members[0].ID)
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

// ImportCustomCertificate calls HttpsCert.ImportCustomCertificate.
// Documentation: https://www.dmtf.org/standards/redfish
func (c *Client) ImportCustomCertificate(ctx context.Context, token, managerPath string, certificate []byte, password string) error {
	_, err := c.do(ctx, resty.MethodPost, managerPath+"/SecurityService/HttpsCert/Actions/HttpsCert.ImportCustomCertificate", token, map[string]string{
		"Certificate": base64.StdEncoding.EncodeToString(certificate),
		"Password":    password,
	}, nil)
	return err
}

// ResetManager calls Manager.Reset with ForceRestart.
// Documentation: https://www.dmtf.org/standards/redfish
func (c *Client) ResetManager(ctx context.Context, token, managerPath string) error {
	_, err := c.do(ctx, resty.MethodPost, managerPath+"/Actions/Manager.Reset", token, map[string]string{
		"ResetType": "ForceRestart",
	}, nil)
	return err
}

func (c *Client) absoluteURL(location string) string {
	u, err := url.Parse(location)
	if err != nil || u.IsAbs() {
		return location
	}
	base, err := url.Parse(c.rc.BaseURL)
	if err != nil {
		return location
	}
	return base.ResolveReference(u).String()
}

func (c *Client) do(ctx context.Context, method, path, token string, payload, result any) (*resty.Response, error) {
	request := c.rc.R().SetContext(ctx)
	if token != "" {
		request.SetHeader("X-Auth-Token", token)
	}
	if payload != nil {
		request.SetHeader("Content-Type", "application/json").SetBody(payload)
	}
	response, err := request.Execute(method, path)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		var apiErr redfishError
		if json.Unmarshal(response.Body(), &apiErr) == nil && (apiErr.Error.Code != "" || apiErr.Error.Message != "") {
			return nil, fmt.Errorf("HTTP %d: %s: %s", response.StatusCode(), apiErr.Error.Code, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", response.StatusCode(), strings.TrimSpace(response.String()))
	}
	if result != nil && len(response.Body()) > 0 {
		if err := json.Unmarshal(response.Body(), result); err != nil {
			return nil, fmt.Errorf("invalid iBMC response: %w", err)
		}
	}
	return response, nil
}
