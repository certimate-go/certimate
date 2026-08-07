// A mock HTTP client for uniCloud.
package unicloud

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

type Client struct {
	username string
	password string

	serverlessToken   string
	serverlessTokenAt time.Time
	serverlessTokenMu sync.Mutex

	apiUserToken   string
	apiUserTokenMu sync.Mutex

	rcForServerless *resty.Client
	rcForApiUser    *resty.Client
}

const (
	uniIdentityEndpoint     = "https://account.dcloud.net.cn/client"
	uniIdentityClientSecret = "ba461799-fde8-429f-8cc4-4b6d306e2339"
	uniIdentityAppId        = "__UNI__uniid_server"
	uniIdentitySpaceId      = "uni-id-server"
	uniConsoleEndpoint      = "https://unicloud.dcloud.net.cn/client"
	uniConsoleClientSecret  = "4c1f7fbf-c732-42b0-ab10-4634a8bbe834"
	uniConsoleAppId         = "__UNI__unicloud_console"
	uniConsoleSpaceId       = "dc-6nfabcn6ada8d3dd"
)

func NewClient(optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.Username == "" {
		return nil, fmt.Errorf("sdkerr: unset username")
	}
	if opts.Password == "" {
		return nil, fmt.Errorf("sdkerr: unset password")
	}

	client := &Client{
		username: opts.Username,
		password: opts.Password,
	}
	client.rcForServerless = resty.New().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent)
	client.rcForApiUser = resty.New().
		SetBaseURL("https://unicloud-api.dcloud.net.cn/unicloud/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(_ *resty.Client, req *http.Request) error {
			if client.apiUserToken != "" {
				req.Header.Set("Token", client.apiUserToken)
			}

			return nil
		})

	return client, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rcForServerless.SetTimeout(timeout)
	return c
}

func (c *Client) buildServerlessClientInfo(appId string) (_clientInfo map[string]any, _err error) {
	return map[string]any{
		"PLATFORM":           "web",
		"OS":                 strings.ToUpper(runtime.GOOS),
		"APPID":              appId,
		"DEVICEID":           app.AppName,
		"LOCALE":             "zh-Hans",
		"osName":             runtime.GOOS,
		"appId":              appId,
		"appName":            "uniCloud",
		"deviceId":           app.AppName,
		"deviceType":         "pc",
		"uniPlatform":        "web",
		"uniCompilerVersion": "4.45",
		"uniRuntimeVersion":  "4.45",
	}, nil
}

func (c *Client) buildServerlessPayloadInfo(appId, spaceId, target, method, action string, params, data any) (map[string]any, error) {
	clientInfo, err := c.buildServerlessClientInfo(appId)
	if err != nil {
		return nil, err
	}

	functionArgsParams := make([]any, 0)
	if params != nil {
		functionArgsParams = append(functionArgsParams, params)
	}

	functionArgs := map[string]any{
		"clientInfo": clientInfo,
		"uniIdToken": c.serverlessToken,
	}
	if method != "" {
		functionArgs["method"] = method
		functionArgs["params"] = make([]any, 0)
	}
	if action != "" {
		type _obj struct{}
		functionArgs["action"] = action
		functionArgs["data"] = &_obj{}
	}
	if params != nil {
		functionArgs["params"] = []any{params}
	}
	if data != nil {
		functionArgs["data"] = data
	}

	jsonb, err := json.Marshal(map[string]any{
		"functionTarget": target,
		"functionArgs":   functionArgs,
	})
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"method":    "serverless.function.runtime.invoke",
		"params":    string(jsonb),
		"spaceId":   spaceId,
		"timestamp": time.Now().UnixMilli(),
	}

	return payload, nil
}

func (c *Client) invokeServerless(ctx context.Context, endpoint, clientSecret, appId, spaceId, target, method, action string, params, data any) (*resty.Response, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("sdkerr: bad request: endpoint cannot be empty")
	}

	payload, err := c.buildServerlessPayloadInfo(appId, spaceId, target, method, action, params, data)
	if err != nil {
		return nil, fmt.Errorf("sdkerr: bad request: failed to build request: %w", err)
	}

	clientInfo, _ := c.buildServerlessClientInfo(appId)
	clientInfoJsonb, _ := json.Marshal(clientInfo)

	sign := generateSignature(payload, clientSecret)

	req := c.rcForServerless.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Origin", "https://unicloud.dcloud.net.cn").
		SetHeader("Referer", "https://unicloud.dcloud.net.cn").
		SetHeader("X-Client-Info", string(clientInfoJsonb)).
		SetHeader("X-Client-Token", c.serverlessToken).
		SetHeader("X-Serverless-Sign", sign).
		SetBody(payload).
		SetContext(ctx)
	resp, err := req.Post(endpoint)
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) invokeServerlessWithResult(ctx context.Context, endpoint, clientSecret, appId, spaceId, target, method, action string, params, data any, result sdkResponse) error {
	resp, err := c.invokeServerless(ctx, endpoint, clientSecret, appId, spaceId, target, method, action, params, data)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &result)
		}
		return err
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	} else if rSuccess := result.GetSuccess(); !rSuccess {
		return fmt.Errorf("sdkerr: api error: code='%s', message='%s'", result.GetErrorCode(), result.GetErrorMessage())
	}

	return nil
}

func (c *Client) sendRequest(ctx context.Context, method string, path string, params any) (*resty.Response, error) {
	req := c.rcForApiUser.R().
		SetContext(ctx)
	if strings.EqualFold(method, http.MethodGet) {
		qs := make(map[string]string)
		if params != nil {
			temp := make(map[string]any)
			jsonb, _ := json.Marshal(params)
			json.Unmarshal(jsonb, &temp)
			for k, v := range temp {
				if v != nil {
					qs[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		req = req.SetQueryParams(qs)
	} else {
		req = req.SetHeader("Content-Type", "application/json").SetBody(params)
	}

	resp, err := req.Execute(method, path)
	if err != nil {
		return resp, fmt.Errorf("sdkerr: failed to send request: %w", err)
	} else if resp.IsError() {
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", resp.StatusCode(), resp.String())
	}

	return resp, nil
}

func (c *Client) sendRequestWithResult(ctx context.Context, method string, path string, params any, result sdkResponse) error {
	resp, err := c.sendRequest(ctx, method, path, params)
	if err != nil {
		if resp != nil {
			json.Unmarshal(resp.Body(), &result)
		}
		return err
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
	} else if rReturnCode := result.GetReturnCode(); rReturnCode != 0 {
		return fmt.Errorf("sdkerr: ret='%d', desc='%s'", rReturnCode, result.GetReturnDesc())
	}

	return nil
}

func (c *Client) ensureServerlessToken(ctx context.Context) error {
	c.serverlessTokenMu.Lock()
	defer c.serverlessTokenMu.Unlock()
	if c.serverlessToken != "" && c.serverlessTokenAt.After(time.Now()) {
		return nil
	}

	params := map[string]string{
		"password": "password",
	}
	if regexp.MustCompile(`^1\d{10}$`).MatchString(c.username) {
		params["mobile"] = c.username
	} else if regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(c.username) {
		params["email"] = c.username
	} else {
		params["username"] = c.username
	}

	type loginResponse struct {
		sdkResponseBase
		Data *struct {
			Code     int32  `json:"errCode"`
			UID      string `json:"uid"`
			NewToken *struct {
				Token        string `json:"token"`
				TokenExpired int64  `json:"tokenExpired"`
			} `json:"newToken,omitempty"`
		} `json:"data,omitempty"`
	}

	resp := &loginResponse{}
	if err := c.invokeServerlessWithResult(ctx,
		uniIdentityEndpoint, uniIdentityClientSecret, uniIdentityAppId, uniIdentitySpaceId,
		"uni-id-co", "login", "", params, nil,
		resp); err != nil {
		return err
	} else {
		if resp.Data == nil || resp.Data.NewToken == nil || resp.Data.NewToken.Token == "" {
			return fmt.Errorf("sdkerr: auth error: received empty token")
		}

		tokenAt := time.UnixMilli(resp.Data.NewToken.TokenExpired)
		if tokenAt.IsZero() {
			return fmt.Errorf("sdkerr: auth error: received invalid token expiration")
		}

		c.serverlessToken = resp.Data.NewToken.Token
		c.serverlessTokenAt = tokenAt
	}

	return nil
}

func (c *Client) ensureApiUserToken(ctx context.Context) error {
	if err := c.ensureServerlessToken(ctx); err != nil {
		return err
	}

	c.apiUserTokenMu.Lock()
	defer c.apiUserTokenMu.Unlock()
	if c.apiUserToken != "" {
		return nil
	}

	type getUserTokenResponse struct {
		sdkResponseBase
		Data *struct {
			Code int32 `json:"code"`
			Data *struct {
				Result      int32  `json:"ret"`
				Description string `json:"desc"`
				Data        *struct {
					Email string `json:"email"`
					Token string `json:"token"`
				} `json:"data,omitempty"`
			} `json:"data,omitempty"`
		} `json:"data,omitempty"`
	}

	resp := &getUserTokenResponse{}
	if err := c.invokeServerlessWithResult(ctx,
		uniConsoleEndpoint, uniConsoleClientSecret, uniConsoleAppId, uniConsoleSpaceId,
		"uni-cloud-kernel", "", "user/getUserToken", nil, map[string]any{"isLogin": true},
		resp); err != nil {
		return err
	} else {
		if resp.Data == nil || resp.Data.Data == nil || resp.Data.Data.Data == nil || resp.Data.Data.Data.Token == "" {
			return fmt.Errorf("sdkerr: auth error: received empty token")
		}

		c.apiUserToken = resp.Data.Data.Data.Token
	}

	return nil
}

func generateSignature(params map[string]any, secret string) string {
	keys := xmaps.Keys(params)
	sort.Strings(keys)

	canonicalStr := ""
	for i, k := range keys {
		if i > 0 {
			canonicalStr += "&"
		}
		canonicalStr += k + "=" + fmt.Sprintf("%v", params[k])
	}

	mac := hmac.New(md5.New, []byte(secret))
	mac.Write([]byte(canonicalStr))
	sign := mac.Sum(nil)
	signHex := hex.EncodeToString(sign)

	return signHex
}
