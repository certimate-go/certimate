package openapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pocketbase/pocketbase/tools/security"

	"github.com/certimate-go/certimate/internal/app"
)

type Client struct {
	rc *resty.Client
}

func NewClient(baseUrl string, optFns ...OptionsFunc) (*Client, error) {
	opts := &Options{}
	for _, fn := range optFns {
		fn(opts)
	}

	if baseUrl == "" {
		return nil, fmt.Errorf("sdkerr: unset baseUrl")
	}
	if _, err := url.Parse(baseUrl); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid baseUrl: %w", err)
	}
	if opts.AccessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if opts.SecretAccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretAccessKey")
	}

	restyClient := resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", app.AppUserAgent).
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// API 签名机制：
			// https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=77&api=%u6784%u9020%u8BF7%u6C42&data=114&vid=107

			now := time.Now()
			eopDate := now.Format("20060102T150405Z")
			eopReqId := security.RandomString(32)

			queryStr := ""
			if req.URL != nil {
				queryStr = req.URL.Query().Encode()
			}

			payloadStr := ""
			if req.Method != http.MethodGet && req.Body != nil {
				payloadb, err := io.ReadAll(req.Body)
				if err != nil {
					return err
				}

				payloadStr = string(payloadb)
				req.Body = io.NopCloser(bytes.NewReader(payloadb))
			}
			payloadHash := sha256.Sum256([]byte(payloadStr))
			payloadHashHex := hex.EncodeToString(payloadHash[:])

			var h hash.Hash
			h = hmac.New(sha256.New, []byte(opts.SecretAccessKey))
			h.Write([]byte(eopDate))
			kTime := h.Sum(nil)
			h = hmac.New(sha256.New, kTime)
			h.Write([]byte(opts.AccessKeyId))
			kAk := h.Sum(nil)
			h = hmac.New(sha256.New, kAk)
			h.Write([]byte(now.Format("20060102")))
			kDate := h.Sum(nil)

			stringToSign := fmt.Sprintf("ctyun-eop-request-id:%s\neop-date:%s\n\n%s\n%s", eopReqId, eopDate, queryStr, payloadHashHex)

			h = hmac.New(sha256.New, kDate)
			h.Write([]byte(stringToSign))
			signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

			req.Header.Set("ctyun-eop-request-id", eopReqId)
			req.Header.Set("eop-date", eopDate)
			req.Header.Set("eop-authorization", fmt.Sprintf("%s Headers=ctyun-eop-request-id;eop-date Signature=%s", opts.AccessKeyId, signature))

			return nil
		})

	return &Client{rc: restyClient}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.rc.SetTimeout(timeout)
	return c
}

func (c *Client) NewRequest(method string, path string) (*resty.Request, error) {
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

func (c *Client) DoRequest(req *resty.Request) (*resty.Response, error) {
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

func (c *Client) DoRequestWithResult(req *resty.Request, res any) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := c.DoRequest(req)
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
