package f5bigip

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

var ErrNotFound = errors.New("not found")

type Client struct {
	baseUrl string
	token   string
	rc      *resty.Client
}

func NewClient(serverUrl string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}

	c := &Client{
		baseUrl: strings.TrimSuffix(serverUrl, "/"),
	}

	c.rc = resty.New().
		SetBaseURL(c.baseUrl).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if c.token != "" {
				req.Header.Set("X-F5-Auth-Token", c.token)
			}
			return nil
		})

	return c, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.rc.SetTLSClientConfig(config)
	return c
}

func (c *Client) Login(ctx context.Context, username, password string) error {
	loginReq := map[string]string{
		"username":          username,
		"password":          password,
		"loginProviderName": "tmos",
	}

	httpreq, err := c.newRequest(http.MethodPost, "/mgmt/shared/authn/login")
	if err != nil {
		return err
	}
	httpreq.SetBody(loginReq)
	httpreq.SetContext(ctx)

	type loginResponse struct {
		Token struct {
			Token string `json:"token"`
		} `json:"token"`
	}

	result := &loginResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	}

	c.token = result.Token.Token
	return nil
}

func (c *Client) UploadCertificate(ctx context.Context, name, partition, content string) error {
	path := fmt.Sprintf("/mgmt/tm/sys/file/ssl-cert")
	body := map[string]string{
		"name":      name,
		"partition": partition,
		"content":   content,
	}

	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return err
	}
	httpreq.SetBody(body)
	httpreq.SetContext(ctx)

	if _, err := c.doRequest(httpreq); err != nil {
		return fmt.Errorf("failed to upload certificate: %w", err)
	}

	return nil
}

func (c *Client) UploadKey(ctx context.Context, name, partition, content string) error {
	path := fmt.Sprintf("/mgmt/tm/sys/file/ssl-key")
	body := map[string]string{
		"name":      name,
		"partition": partition,
		"content":   content,
	}

	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return err
	}
	httpreq.SetBody(body)
	httpreq.SetContext(ctx)

	if _, err := c.doRequest(httpreq); err != nil {
		return fmt.Errorf("failed to upload key: %w", err)
	}

	return nil
}

func (c *Client) GetClientSSLProfile(ctx context.Context, name, partition string) (map[string]any, error) {
	path := fmt.Sprintf("/mgmt/tm/ltm/profile/client-ssl/~%s~%s", url.PathEscape(partition), url.PathEscape(name))

	httpreq, err := c.newRequest(http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	httpreq.SetContext(ctx)

	result := make(map[string]any)
	resp, err := c.doRequestWithResult(httpreq, &result)
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			return nil, fmt.Errorf("failed to get client-ssl profile: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get client-ssl profile: %w", err)
	}

	return result, nil
}

func (c *Client) CreateClientSSLProfile(ctx context.Context, name, partition, certPath, keyPath, chainPath string) error {
	path := fmt.Sprintf("/mgmt/tm/ltm/profile/client-ssl")
	body := map[string]string{
		"name":          name,
		"partition":     partition,
		"defaultsFrom":  "clientssl",
		"cert":          certPath,
		"key":           keyPath,
	}
	if chainPath != "" {
		body["chain"] = chainPath
	}

	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return err
	}
	httpreq.SetBody(body)
	httpreq.SetContext(ctx)

	if _, err := c.doRequest(httpreq); err != nil {
		return fmt.Errorf("failed to create client-ssl profile: %w", err)
	}

	return nil
}

func (c *Client) UpdateClientSSLProfile(ctx context.Context, name, partition, certPath, keyPath, chainPath string) error {
	path := fmt.Sprintf("/mgmt/tm/ltm/profile/client-ssl/~%s~%s", url.PathEscape(partition), url.PathEscape(name))
	body := map[string]string{
		"cert": certPath,
		"key":  keyPath,
	}
	if chainPath != "" {
		body["chain"] = chainPath
	}

	httpreq, err := c.newRequest(http.MethodPatch, path)
	if err != nil {
		return err
	}
	httpreq.SetBody(body)
	httpreq.SetContext(ctx)

	if _, err := c.doRequest(httpreq); err != nil {
		return fmt.Errorf("failed to update client-ssl profile: %w", err)
	}

	return nil
}

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.rc.R()
	req.Method = method
	req.URL = path
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := req.Send()
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) doRequestWithResult(req *resty.Request, res interface{}) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := c.doRequest(req)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &res)
		}
		return resp, err
	}

	if len(resp.Body()) != 0 {
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		}
	}

	return resp, nil
}
