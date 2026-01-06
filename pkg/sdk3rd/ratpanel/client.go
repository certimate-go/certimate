package ratpanel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	client *resty.Client
}

func NewClient(serverUrl string, accessTokenId int64, accessToken string) (*Client, error) {
	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if accessTokenId == 0 {
		return nil, fmt.Errorf("sdkerr: unset accessTokenId")
	}
	if accessToken == "" {
		return nil, fmt.Errorf("sdkerr: unset accessToken")
	}

	client := resty.New().
		SetBaseURL(strings.TrimRight(serverUrl, "/")+"/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			var body []byte
			var err error

			if req.Body != nil {
				body, err = io.ReadAll(req.Body)
				if err != nil {
					return err
				}
				req.Body = io.NopCloser(bytes.NewReader(body))
			}

			canonicalPath := req.URL.Path
			if !strings.HasPrefix(canonicalPath, "/api") {
				index := strings.Index(canonicalPath, "/api")
				if index != -1 {
					canonicalPath = canonicalPath[index:]
				}
			}

			canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s",
				req.Method,
				canonicalPath,
				req.URL.Query().Encode(),
				sumSha256(string(body)))

			timestamp := time.Now().Unix()
			req.Header.Set("X-Timestamp", fmt.Sprintf("%d", timestamp))

			stringToSign := fmt.Sprintf("%s\n%d\n%s",
				"HMAC-SHA256",
				timestamp,
				sumSha256(canonicalRequest))
			signature := sumHmacSha256(stringToSign, accessToken)
			req.Header.Set("Authorization", fmt.Sprintf("HMAC-SHA256 Credential=%d, Signature=%s", accessTokenId, signature))

			return nil
		})

	return &Client{client}, nil
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

func (c *Client) doRequestWithResult(req *resty.Request, res apiResponse) (*resty.Response, error) {
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
			if tmessage := res.GetMessage(); tmessage != "success" {
				return resp, fmt.Errorf("sdkerr: message='%s'", tmessage)
			}
		}
	}

	return resp, nil
}

func sumSha256(str string) string {
	sum := sha256.Sum256([]byte(str))
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum[:])
	return string(dst)
}

func sumHmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
