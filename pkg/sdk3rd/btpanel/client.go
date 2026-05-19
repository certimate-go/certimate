package btpanel

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	apiKey            string
	serverUrl         string
	allowInsecureCurl bool

	client *resty.Client
}

func NewClient(serverUrl, apiKey string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("sdkerr: unset apiKey")
	}

	client := resty.New().
		SetBaseURL(strings.TrimSuffix(serverUrl, "/")).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("User-Agent", app.AppUserAgent)

	return &Client{
		apiKey:    apiKey,
		serverUrl: strings.TrimSuffix(serverUrl, "/"),
		client:    client,
	}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.client.SetTLSClientConfig(config)
	c.allowInsecureCurl = config != nil && config.InsecureSkipVerify
	return c
}

func (c *Client) newRequest(method string, path string, params any) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("sdkerr: unset method")
	}
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	data := make(map[string]string)
	if params != nil {
		temp := make(map[string]any)
		jsonb, _ := json.Marshal(params)
		json.Unmarshal(jsonb, &temp)
		for k, v := range temp {
			if v == nil {
				continue
			}

			switch reflect.Indirect(reflect.ValueOf(v)).Kind() {
			case reflect.String:
				data[k] = v.(string)

			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				data[k] = fmt.Sprintf("%v", v)

			default:
				if t, ok := v.(time.Time); ok {
					data[k] = t.Format(time.RFC3339)
				} else {
					jsonb, _ := json.Marshal(v)
					data[k] = string(jsonb)
				}
			}
		}
	}

	timestamp := time.Now().Unix()
	data["request_time"] = fmt.Sprintf("%d", timestamp)
	data["request_token"] = generateSignature(fmt.Sprintf("%d", timestamp), c.apiKey)

	req := c.client.R()
	req.Method = method
	req.URL = path
	req.SetFormData(data)
	return req, nil
}

func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	// WARN:
	//   PLEASE DO NOT USE `req.SetBody` or `req.SetFormData` HERE! USE `newRequest` INSTEAD.
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
		if c.shouldRetryWithCurl(err) {
			body, statusCode, curlErr := c.doRequestWithCurl(req)
			if curlErr == nil {
				slog.Warn("btpanel sdk request retried with curl fallback", slog.String("method", req.Method), slog.String("url", req.URL), slog.Any("error", err))
				return nil, c.decodeCurlResponse(body, statusCode, res)
			}
			slog.Warn("btpanel sdk request curl fallback failed", slog.String("method", req.Method), slog.String("url", req.URL), slog.Any("error", err), slog.Any("fallbackError", curlErr))
		}
		return resp, err
	}

	if len(resp.Body()) != 0 {
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, resp.String())
		} else {
			if tstatus := res.GetStatus(); tstatus != nil && !*tstatus {
				if res.GetMessage() == nil {
					return resp, fmt.Errorf("sdkerr: api error: unknown error")
				} else {
					return resp, fmt.Errorf("sdkerr: api error: message='%s'", *res.GetMessage())
				}
			}
		}
	}

	return resp, nil
}

func (c *Client) shouldRetryWithCurl(err error) bool {
	if err == nil || !c.allowInsecureCurl {
		return false
	}

	errmsg := err.Error()
	return strings.Contains(errmsg, "tls: handshake failure") ||
		strings.Contains(errmsg, "server gave HTTP response to HTTPS client") ||
		strings.Contains(errmsg, "first record does not look like a TLS handshake")
}

func (c *Client) doRequestWithCurl(req *resty.Request) ([]byte, int, error) {
	if _, err := exec.LookPath("curl"); err != nil {
		return nil, 0, err
	}

	requestUrl := req.URL
	if !strings.HasPrefix(requestUrl, "http://") && !strings.HasPrefix(requestUrl, "https://") {
		requestUrl = c.serverUrl + "/" + strings.TrimPrefix(req.URL, "/")
	}
	if len(req.QueryParam) > 0 {
		requestUrl += "?" + req.QueryParam.Encode()
	}

	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	args := []string{
		"-sS",
		"-k",
		"-X", req.Method,
		"-H", "Accept: application/json",
		"-H", "Content-Type: application/x-www-form-urlencoded",
		"-H", "User-Agent: " + app.AppUserAgent,
		"--data-binary", "@-",
		"-w", "\n%{http_code}",
		requestUrl,
	}
	cmd := exec.CommandContext(ctx, "curl", args...)
	cmd.Stdin = strings.NewReader(req.FormData.Encode())

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, 0, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, 0, err
	}

	body, statusCode, err := splitCurlOutput(output)
	if err != nil {
		return nil, 0, err
	}

	return body, statusCode, nil
}

func splitCurlOutput(output []byte) ([]byte, int, error) {
	idx := bytes.LastIndexByte(output, '\n')
	if idx < 0 {
		return nil, 0, fmt.Errorf("sdkerr: failed to parse curl response")
	}

	statusCode, err := strconv.Atoi(strings.TrimSpace(string(output[idx+1:])))
	if err != nil {
		return nil, 0, fmt.Errorf("sdkerr: failed to parse curl status code: %w", err)
	}

	return output[:idx], statusCode, nil
}

func (c *Client) decodeCurlResponse(body []byte, statusCode int, res sdkResponse) error {
	if statusCode >= http.StatusBadRequest {
		return fmt.Errorf("sdkerr: unexpected status code: %d (resp: %s)", statusCode, string(body))
	}

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("sdkerr: failed to unmarshal response: %w (resp: %s)", err, string(body))
	}

	if tstatus := res.GetStatus(); tstatus != nil && !*tstatus {
		if res.GetMessage() == nil {
			return fmt.Errorf("sdkerr: api error: unknown error")
		}
		return fmt.Errorf("sdkerr: api error: message='%s'", *res.GetMessage())
	}

	return nil
}

func generateSignature(timestamp string, apiKey string) string {
	keyMd5 := md5.Sum([]byte(apiKey))
	keyMd5Hex := strings.ToLower(hex.EncodeToString(keyMd5[:]))

	signMd5 := md5.Sum([]byte(timestamp + keyMd5Hex))
	signMd5Hex := strings.ToLower(hex.EncodeToString(signMd5[:]))
	return signMd5Hex
}
