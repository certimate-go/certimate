package xinnet

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient(agentId, appSecret string) (*Client, error) {
	if agentId == "" {
		return nil, fmt.Errorf("sdkerr: unset agentId")
	}
	if appSecret == "" {
		return nil, fmt.Errorf("sdkerr: unset appSecret")
	}

	client := resty.New().
		SetBaseURL("https://apiv2.xinnet.com/api").
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "certimate").
		SetPreRequestHook(func(c *resty.Client, req *http.Request) error {
			// 生成时间戳
			timestamp := time.Now().UTC().Format("20060102T150405Z")

			// 获取请求路径，注意结尾必须是 "/"
			urlPath := "/"
			if req.URL != nil {
				urlPath = req.URL.Path

				if !strings.HasSuffix(urlPath, "/") {
					urlPath += "/"
				}
			}

			// 获取请求方法
			requestMethod := req.Method

			// 获取请求体
			requestBody := ""
			if req.Body != nil {
				reader, err := req.GetBody()
				if err != nil {
					return err
				}

				defer reader.Close()

				payloadb, err := io.ReadAll(reader)
				if err != nil {
					return err
				}

				requestBody = string(payloadb)
			}

			// 计算签名
			algorithm := "HMAC-SHA256"
			stringToSign := algorithm + "\n" +
				timestamp + "\n" +
				requestMethod + "\n" +
				urlPath + "\n" +
				requestBody
			h := hmac.New(sha256.New, []byte(appSecret))
			h.Write([]byte(stringToSign))
			signature := hex.EncodeToString(h.Sum(nil))

			// 设置请求头
			req.Header.Set("timestamp", timestamp)
			req.Header.Set("authorization", fmt.Sprintf("%s Access=%s, Signature=%s", algorithm, agentId, signature))

			return nil
		})

	return &Client{client}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) newRequest(path string) (*resty.Request, error) {
	if path == "" {
		return nil, fmt.Errorf("sdkerr: unset path")
	}

	req := c.client.R()
	req.Method = http.MethodPost
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
			if tcode := res.GetCode(); tcode != "0" {
				return resp, fmt.Errorf("sdkerr: code='%s', msg='%s'", tcode, res.GetMessage())
			}
		}
	}

	return resp, nil
}
