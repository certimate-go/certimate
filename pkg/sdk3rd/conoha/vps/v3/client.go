// A simple SDK client for ConoHa VPS v3.
// API documentation: https://doc.conoha.jp/reference/api-vps3/
package v3

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	userId       string
	userName     string
	userPassword string
	tenantId     string
	tenantName   string

	token   string
	tokenAt time.Time
	tokenMu sync.Mutex

	rc *resty.Client
}

func NewClient(optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.UserId == "" && opts.UserName == "" {
		return nil, fmt.Errorf("sdkerr: unset userId or userName")
	}
	if opts.UserPassword == "" {
		return nil, fmt.Errorf("sdkerr: unset userPassword")
	}
	if opts.TenantId == "" || opts.TenantName == "" {
		return nil, fmt.Errorf("sdkerr: unset tenantId or tenantName")
	}

	client := &Client{
		userId:       opts.UserId,
		userName:     opts.UserName,
		userPassword: opts.UserPassword,
		tenantId:     opts.TenantId,
		tenantName:   opts.TenantName,
	}
	client.rc = resty.New().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if client.token != "" {
				req.Header.Set("X-Auth-Token", client.token)
			}

			return nil
		})

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.rc.SetTLSClientConfig(config)
	return c
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

	// WARN:
	//   DO NOT CALL `req.SetResult` or `req.SetError` AGAIN! USE `doRequestWithResult` INSTEAD.
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

func (c *Client) doRequestWithResult(req *resty.Request, res sdkResponse) (*resty.Response, error) {
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
		} else {
			if rCode := res.GetCode(); rCode == 0 {
				return resp, fmt.Errorf("sdkerr: api error: code='%d', error='%s'", rCode, res.GetError())
			}
		}
	}

	return resp, nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && c.tokenAt.After(time.Now()) {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, identityBaseURL+"/v3/auth/tokens")
	if err != nil {
		return err
	} else {
		authUserParams := map[string]string{"password": c.userPassword}
		if c.userId != "" {
			authUserParams["id"] = c.userId
		}
		if c.userName != "" {
			authUserParams["name"] = c.userName
		}

		authProjectParams := map[string]string{}
		if c.tenantId != "" {
			authProjectParams["id"] = c.tenantId
		}
		if c.tenantName != "" {
			authProjectParams["name"] = c.tenantName
		}

		httpreq.SetBody(map[string]any{
			"auth": map[string]any{
				"identity": map[string]any{
					"methods": []string{"password"},
					"password": map[string]any{
						"user": authUserParams,
					},
					"scope": map[string]any{
						"project": authProjectParams,
					},
				},
			},
		})
		httpreq.SetContext(ctx)
	}

	type createAuthTokenResponse struct {
		sdkResponseBase
		Token *struct {
			IssuedAt  string `json:"issued_at"`
			ExpiresAt string `json:"expires_at"`
		} `json:"token,omitempty"`
	}

	result := &createAuthTokenResponse{}
	if httpresp, err := c.doRequestWithResult(httpreq, result); err != nil {
		return err
	} else if rCode := result.GetCode(); rCode != 0 {
		return fmt.Errorf("sdkerr: auth error: code='%d', error='%s'", rCode, result.GetError())
	} else {
		token := httpresp.Header().Get("X-Subject-Token")
		if token == "" {
			return fmt.Errorf("sdkerr: auth error: received empty token")
		}

		tokenAt, err := time.Parse(time.RFC3339Nano, result.Token.ExpiresAt)
		if err != nil {
			return fmt.Errorf("sdkerr: auth error: received invalid token expiration: %w", err)
		}

		c.token = token
		c.tokenAt = tokenAt
	}

	return nil
}
