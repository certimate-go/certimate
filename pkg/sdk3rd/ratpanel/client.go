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
	rc *resty.Client
}

func NewClient(serverUrl string, optFns ...OptionsFunc) (*Client, error) {
	options := &Options{}
	for _, fn := range optFns {
		fn(options)
	}

	if serverUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset serverUrl")
	}
	if _, err := url.Parse(serverUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid serverUrl: %w", err)
	}
	if options.AccessTokenId == 0 {
		return nil, fmt.Errorf("sdkerr: unset accessTokenId")
	}
	if options.AccessToken == "" {
		return nil, fmt.Errorf("sdkerr: unset accessToken")
	}

	restyClient := resty.New().
		SetBaseURL(strings.TrimSuffix(serverUrl, "/")+"/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// API 签名机制：
			// https://ratpanel.github.io/advanced/api#authentication-mechanism

			payloadStr := ""
			if req.Body != nil {
				payloadb, err := io.ReadAll(req.Body)
				if err != nil {
					return err
				}

				payloadStr = string(payloadb)
				req.Body = io.NopCloser(bytes.NewReader(payloadb))
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
				sumSha256(payloadStr),
			)

			timestamp := time.Now().Unix()

			stringToSign := fmt.Sprintf("%s\n%d\n%s",
				"HMAC-SHA256",
				timestamp,
				sumSha256(canonicalRequest),
			)

			signature := sumHmacSha256(stringToSign, options.AccessToken)

			req.Header.Set("X-Timestamp", fmt.Sprintf("%d", timestamp))
			req.Header.Set("Authorization", fmt.Sprintf("HMAC-SHA256 Credential=%d, Signature=%s", options.AccessTokenId, signature))

			return nil
		})

	return &Client{rc: restyClient}, nil
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
			if rMessage := res.GetMessage(); rMessage != "success" {
				return resp, fmt.Errorf("sdkerr: message='%s'", rMessage)
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
