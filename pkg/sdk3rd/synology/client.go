package synology

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"

	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
)

// Client is a Synology DSM API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	sid        string
	synoToken  string
	apiPath    string
	apiVersion int
}

// NewClient creates a new Synology DSM API client.
func NewClient(hostname string, port int, scheme string, allowInsecure bool) *Client {
	if scheme == "" {
		scheme = "http"
	}
	if port == 0 {
		if scheme == "https" {
			port = 5001
		} else {
			port = 5000
		}
	}

	baseURL := fmt.Sprintf("%s://%s:%d", scheme, hostname, port)

	httpClient := &http.Client{
		Transport: xhttp.NewDefaultTransport(),
		Timeout:   http.DefaultClient.Timeout,
	}
	if allowInsecure {
		transport := xhttp.NewDefaultTransport()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = transport
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Login authenticates with Synology DSM and gets session ID and SynoToken.
func (c *Client) Login(username, password, otpCode string) error {
	// Step 1: Query API info to get path and version
	apiInfo, err := c.queryAPIInfo()
	if err != nil {
		return fmt.Errorf("failed to query API info: %w", err)
	}

	authInfo, ok := apiInfo["SYNO.API.Auth"]
	if !ok {
		return fmt.Errorf("SYNO.API.Auth not found in API info")
	}

	c.apiPath = authInfo.Path
	c.apiVersion = authInfo.MaxVersion

	// Step 2: Login
	params := url.Values{
		"api":               {"SYNO.API.Auth"},
		"version":           {strconv.Itoa(c.apiVersion)},
		"method":            {"login"},
		"format":            {"sid"},
		"account":           {username},
		"passwd":            {password},
		"enable_syno_token": {"yes"},
	}

	if otpCode != "" {
		params.Set("otp_code", otpCode)
		params.Set("enable_device_token", "yes")
		params.Set("device_name", "Certimate")
	}

	loginURL := fmt.Sprintf("%s/webapi/%s?%s", c.baseURL, c.apiPath, params.Encode())

	resp, err := c.httpClient.Get(loginURL)
	if err != nil {
		return fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if !loginResp.Success {
		if loginResp.Error != nil {
			return fmt.Errorf("login failed: %s (error code: %d)", getAuthErrorDescription(loginResp.Error.Code), loginResp.Error.Code)
		}
		return fmt.Errorf("login failed: unknown error")
	}

	c.sid = loginResp.Data.Sid
	c.synoToken = loginResp.Data.SynoToken

	if c.sid == "" || c.synoToken == "" {
		return fmt.Errorf("login succeeded but session ID or SynoToken is empty")
	}

	return nil
}

// Logout ends the session with Synology DSM.
func (c *Client) Logout() error {
	if c.sid == "" {
		return nil
	}

	params := url.Values{
		"api":     {"SYNO.API.Auth"},
		"version": {strconv.Itoa(c.apiVersion)},
		"method":  {"logout"},
		"_sid":    {c.sid},
	}

	logoutURL := fmt.Sprintf("%s/webapi/%s?%s", c.baseURL, c.apiPath, params.Encode())

	resp, err := c.httpClient.Get(logoutURL)
	if err != nil {
		return fmt.Errorf("failed to send logout request: %w", err)
	}
	defer resp.Body.Close()

	c.sid = ""
	c.synoToken = ""

	return nil
}

// ListCertificates retrieves the list of certificates from Synology DSM.
func (c *Client) ListCertificates() ([]Certificate, error) {
	params := url.Values{
		"api":     {"SYNO.Core.Certificate.CRT"},
		"method":  {"list"},
		"version": {"1"},
		"_sid":    {c.sid},
	}

	reqURL := fmt.Sprintf("%s/webapi/entry.cgi", c.baseURL)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create list certificates request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-SYNO-TOKEN", c.synoToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send list certificates request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read list certificates response: %w", err)
	}

	var listResp ListCertificatesResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse list certificates response: %w", err)
	}

	if !listResp.Success {
		if listResp.Error != nil {
			return nil, fmt.Errorf("list certificates failed with error code: %d", listResp.Error.Code)
		}
		return nil, fmt.Errorf("list certificates failed: unknown error")
	}

	return listResp.Data.Certificates, nil
}

// ImportCertificate imports or updates a certificate in Synology DSM.
func (c *Client) ImportCertificate(certPEM, keyPEM, caPEM, certID, description string, isDefault bool) error {
	// Prepare multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add key file (must be first as per acme.sh)
	keyWriter, err := writer.CreateFormFile("key", "privkey.pem")
	if err != nil {
		return fmt.Errorf("failed to create key form field: %w", err)
	}
	if _, err := keyWriter.Write([]byte(keyPEM)); err != nil {
		return fmt.Errorf("failed to write key data: %w", err)
	}

	// Add cert file (server certificate only)
	certWriter, err := writer.CreateFormFile("cert", "cert.pem")
	if err != nil {
		return fmt.Errorf("failed to create cert form field: %w", err)
	}
	if _, err := certWriter.Write([]byte(certPEM)); err != nil {
		return fmt.Errorf("failed to write cert data: %w", err)
	}

	// Add intermediate cert (CA) file - always include even if empty
	caWriter, err := writer.CreateFormFile("inter_cert", "chain.pem")
	if err != nil {
		return fmt.Errorf("failed to create inter_cert form field: %w", err)
	}
	if _, err := caWriter.Write([]byte(caPEM)); err != nil {
		return fmt.Errorf("failed to write inter_cert data: %w", err)
	}

	// Add certificate ID - always send even if empty (required by DSM API)
	if err := writer.WriteField("id", certID); err != nil {
		return fmt.Errorf("failed to write id field: %w", err)
	}

	// Add description - always send even if empty (required by DSM API)
	if err := writer.WriteField("desc", description); err != nil {
		return fmt.Errorf("failed to write desc field: %w", err)
	}

	// Add as_default flag
	if isDefault {
		if err := writer.WriteField("as_default", "true"); err != nil {
			return fmt.Errorf("failed to write as_default field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build request URL
	params := url.Values{
		"api":       {"SYNO.Core.Certificate"},
		"method":    {"import"},
		"version":   {"1"},
		"SynoToken": {c.synoToken},
		"_sid":      {c.sid},
	}
	reqURL := fmt.Sprintf("%s/webapi/entry.cgi?%s", c.baseURL, params.Encode())

	// Create request
	req, err := http.NewRequest(http.MethodPost, reqURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create import request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-SYNO-TOKEN", c.synoToken)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send import request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read import response: %w", err)
	}

	var importResp ImportCertificateResponse
	if err := json.Unmarshal(body, &importResp); err != nil {
		return fmt.Errorf("failed to parse import response: %w", err)
	}

	if !importResp.Success {
		if importResp.Error != nil {
			return fmt.Errorf("import certificate failed with error code: %d", importResp.Error.Code)
		}
		return fmt.Errorf("import certificate failed: unknown error")
	}

	return nil
}

// queryAPIInfo queries the API info from Synology DSM.
func (c *Client) queryAPIInfo() (map[string]APIInfo, error) {
	queryURL := fmt.Sprintf("%s/webapi/query.cgi?api=SYNO.API.Info&version=1&method=query&query=SYNO.API.Auth", c.baseURL)

	resp, err := c.httpClient.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send query request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read query response: %w", err)
	}

	var queryResp QueryAPIInfoResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	if !queryResp.Success {
		return nil, fmt.Errorf("query API info failed")
	}

	return queryResp.Data, nil
}

// getAuthErrorDescription returns a human-readable description for Synology DSM auth error codes.
// Reference: https://global.synologydownload.com/download/Document/Software/DeveloperGuide/Os/DSM/All/enu/Synology_DiskStation_Manager_API_Guide.pdf
func getAuthErrorDescription(code int) string {
	switch code {
	case 100:
		return "Unknown error"
	case 101:
		return "Invalid parameters"
	case 102:
		return "API does not exist"
	case 103:
		return "Method does not exist"
	case 104:
		return "This API version is not supported"
	case 105:
		return "Insufficient user privilege"
	case 106:
		return "Connection time out"
	case 107:
		return "Multiple login detected"
	case 400:
		return "Invalid password or account does not exist"
	case 401:
		return "Guest or disabled account"
	case 402:
		return "Permission denied"
	case 403:
		return "2-factor authentication code required (OTP)"
	case 404:
		return "Failed to authenticate 2-factor authentication code"
	case 405:
		return "Server version is too low or not supported"
	case 406:
		return "2-factor authentication code expired"
	case 407:
		return "Login failed: IP has been blocked"
	case 408:
		return "Expired password"
	case 409:
		return "Password must be changed (password policy)"
	case 410:
		return "Account locked (too many failed login attempts)"
	default:
		return "Unknown authentication error"
	}
}

// ListCertificatesWithServices retrieves the list of certificates with their associated services.
func (c *Client) ListCertificatesWithServices() ([]CertificateWithServices, error) {
	params := url.Values{
		"api":     {"SYNO.Core.Certificate.CRT"},
		"method":  {"list"},
		"version": {"1"},
		"_sid":    {c.sid},
	}

	reqURL := fmt.Sprintf("%s/webapi/entry.cgi", c.baseURL)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create list certificates request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-SYNO-TOKEN", c.synoToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send list certificates request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read list certificates response: %w", err)
	}

	var listResp ListCertificatesWithServicesResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse list certificates response: %w", err)
	}

	if !listResp.Success {
		if listResp.Error != nil {
			return nil, fmt.Errorf("list certificates failed with error code: %d", listResp.Error.Code)
		}
		return nil, fmt.Errorf("list certificates failed: unknown error")
	}

	return listResp.Data.Certificates, nil
}

// SetCertificateForAllServices applies the new certificate to all services that were using the old certificate.
// If oldCertID is empty, it will apply the new certificate to ALL services.
func (c *Client) SetCertificateForAllServices(newCertID string, oldCertID string) error {
	// Get all certificates with their services
	certs, err := c.ListCertificatesWithServices()
	if err != nil {
		return fmt.Errorf("failed to list certificates with services: %w", err)
	}

	// Build the settings array - collect all services that need to be updated
	var settings []ServiceCertificateSetting

	for _, cert := range certs {
		// If oldCertID is specified, only update services that use that certificate
		// Otherwise, update all services to use the new certificate
		if oldCertID != "" && cert.ID != oldCertID {
			continue
		}

		// Skip if this is already the new certificate
		if cert.ID == newCertID {
			continue
		}

		for _, service := range cert.Services {
			settings = append(settings, ServiceCertificateSetting{
				Service: service,
				OldID:   cert.ID,
				ID:      newCertID,
			})
		}
	}

	if len(settings) == 0 {
		return nil
	}

	// Convert settings to JSON
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Build request
	data := url.Values{
		"api":      {"SYNO.Core.Certificate.Service"},
		"method":   {"set"},
		"version":  {"1"},
		"settings": {string(settingsJSON)},
	}

	reqURL := fmt.Sprintf("%s/webapi/entry.cgi?_sid=%s", c.baseURL, c.sid)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create set service request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-SYNO-TOKEN", c.synoToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send set service request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read set service response: %w", err)
	}

	var setResp SetServiceCertificateResponse
	if err := json.Unmarshal(body, &setResp); err != nil {
		return fmt.Errorf("failed to parse set service response: %w", err)
	}

	if !setResp.Success {
		if setResp.Error != nil {
			return fmt.Errorf("set service certificate failed with error code: %d", setResp.Error.Code)
		}
		return fmt.Errorf("set service certificate failed: unknown error")
	}

	return nil
}

// GenerateTOTPCode generates a TOTP code from the given secret key.
// The secret should be a base32-encoded string (the format shown in authenticator apps).
func GenerateTOTPCode(secret string) (string, error) {
	// Clean up the secret - remove spaces and convert to uppercase
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))

	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	return code, nil
}
