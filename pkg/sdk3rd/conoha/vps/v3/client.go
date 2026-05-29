package v3

import (
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

	accessToken    string
	accessTokenExp time.Time
	accessTokenMtx sync.Mutex

	client *resty.Client
}

func NewClient(userId, userName, userPassword, tenantId, tenantName string) (*Client, error) {
	if userId == "" && userName == "" {
		return nil, fmt.Errorf("sdkerr: unset userId or userName")
	}
	if userPassword == "" {
		return nil, fmt.Errorf("sdkerr: unset userPassword")
	}
	if tenantId == "" || tenantName == "" {
		return nil, fmt.Errorf("sdkerr: unset tenantId or tenantName")
	}

	client := &Client{
		userId:       userId,
		userName:     userName,
		userPassword: userPassword,
		tenantId:     tenantId,
		tenantName:   tenantName,
	}
	client.client = resty.New().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			if client.accessToken != "" {
				req.Header.Set("X-Auth-Token", client.accessToken)
			}

			return nil
		})

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.client.SetTLSClientConfig(config)
	return c
}

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.client.R()
	req.Method = method
	req.URL = path
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	// WARN:
	//   PLEASE DO NOT USE `req.SetResult` or `req.SetError` HERE! USE `doRequestWithResult` INSTEAD.

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
			if tcode := res.GetCode(); tcode == 0 {
				return resp, fmt.Errorf("sdkerr: code='%d', error='%s'", tcode, res.GetError())
			}
		}
	}

	return resp, nil
}

func (c *Client) ensureAccessTokenExists() error {
	c.accessTokenMtx.Lock()
	defer c.accessTokenMtx.Unlock()
	if c.accessToken != "" && c.accessTokenExp.After(time.Now()) {
		return nil
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("%s/v3/auth/tokens", identityBaseURL))
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
	} else if code := result.GetCode(); code != 0 {
		return fmt.Errorf("sdkerr: failed to get conoha access token: code='%d', error='%s'", code, result.GetError())
	} else {
		token := httpresp.Header().Get("X-Subject-Token")
		if token == "" {
			return fmt.Errorf("sdkerr: api error: received empty auth token")
		}

		tokenExp, err := time.Parse(time.RFC3339Nano, result.Token.ExpiresAt)
		if err != nil {
			return fmt.Errorf("sdkerr: api error: received invalid auth token expiration: %w", err)
		}

		c.accessToken = token
		c.accessTokenExp = tokenExp
	}

	return nil
}
