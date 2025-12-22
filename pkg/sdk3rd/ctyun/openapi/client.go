package openapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pocketbase/pocketbase/tools/security"
)

type Client struct {
	client *resty.Client
}

func NewClient(endpoint, accessKeyId, secretAccessKey string) (*Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("sdkerr: unset endpoint")
	}
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("sdkerr: invalid endpoint: %w", err)
	}
	if accessKeyId == "" {
		return nil, fmt.Errorf("sdkerr: unset accessKeyId")
	}
	if secretAccessKey == "" {
		return nil, fmt.Errorf("sdkerr: unset secretAccessKey")
	}

	client := resty.New().
		SetBaseURL(endpoint).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "certimate").
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// 生成时间戳及流水号
			now := time.Now()
			eopDate := now.Format("20060102T150405Z")
			eopReqId := security.RandomString(32)

			// 获取查询参数
			queryStr := ""
			if req.URL != nil {
				queryStr = req.URL.Query().Encode()
			}

			// 获取请求正文
			payloadStr := ""
			if req.Body != nil {
				reader, err := req.GetBody()
				if err != nil {
					return err
				}

				defer reader.Close()
				payload, err := io.ReadAll(reader)
				if err != nil {
					return err
				}

				payloadStr = string(payload)
			}

			// 构造代签字符串
			payloadHash := sha256.Sum256([]byte(payloadStr))
			payloadHashHex := hex.EncodeToString(payloadHash[:])
			dataToSign := fmt.Sprintf("ctyun-eop-request-id:%s\neop-date:%s\n\n%s\n%s", eopReqId, eopDate, queryStr, payloadHashHex)

			// 生成 ktime
			hasher := hmac.New(sha256.New, []byte(secretAccessKey))
			hasher.Write([]byte(eopDate))
			ktime := hasher.Sum(nil)

			// 生成 kak
			hasher = hmac.New(sha256.New, ktime)
			hasher.Write([]byte(accessKeyId))
			kak := hasher.Sum(nil)

			// 生成 kdate
			hasher = hmac.New(sha256.New, kak)
			hasher.Write([]byte(now.Format("20060102")))
			kdate := hasher.Sum(nil)

			// 构造签名
			hasher = hmac.New(sha256.New, kdate)
			hasher.Write([]byte(dataToSign))
			sign := hasher.Sum(nil)
			signStr := base64.StdEncoding.EncodeToString(sign)

			// 设置请求头
			req.Header.Set("ctyun-eop-request-id", eopReqId)
			req.Header.Set("eop-date", eopDate)
			req.Header.Set("eop-authorization", fmt.Sprintf("%s Headers=ctyun-eop-request-id;eop-date Signature=%s", accessKeyId, signStr))

			return nil
		})

	return &Client{client}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) NewRequest(method string, path string) (*resty.Request, error) {
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
