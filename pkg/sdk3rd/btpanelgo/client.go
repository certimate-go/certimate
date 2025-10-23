package btpanel

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	apiKey string

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
		SetBaseURL(strings.TrimRight(serverUrl, "/")).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("User-Agent", "certimate")

	return &Client{
		apiKey: apiKey,
		client: client,
	}, nil
}

func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.client.SetTimeout(timeout)
	return c
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.client.SetTLSClientConfig(config)
	return c
}

func (c *Client) newRequest(method string, path string, params any, multipart bool) (*resty.Request, error) {
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
	if multipart {
		req.SetMultipartFormData(data)

		if params != nil {
			vparams := reflect.ValueOf(params)
			if vparams.Kind() == reflect.Ptr {
				vparams = vparams.Elem()
			}
			if vparams.Kind() == reflect.Struct {
				vparamsTyp := vparams.Type()
				for i := 0; i < vparams.NumField(); i++ {
					field := vparamsTyp.Field(i)
					fieldVal := vparams.Field(i)
					if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
						continue
					}

					formTag := field.Tag.Get("form")
					if formTag == "" {
						continue
					}

					switch v := fieldVal.Interface().(type) {
					case []byte:
						req.SetMultipartField(formTag, formTag, "", bytes.NewReader(v))
					default:
						panic("unreachable")
					}
				}
			} else {
				panic("unreachable")
			}
		}
	} else {
		req.SetFormData(data)
	}

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
		return resp, fmt.Errorf("sdkerr: unexpected status code: %d, resp: %s", resp.StatusCode(), resp.String())
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
			return resp, fmt.Errorf("sdkerr: failed to unmarshal response: %w", err)
		} else {
			if tstatus := res.GetStatus(); tstatus != nil {
				var errored bool

				var bstatus bool
				if err := json.Unmarshal(tstatus, &bstatus); err == nil {
					errored = !bstatus
				}

				var istatus int
				if err := json.Unmarshal(tstatus, &istatus); err == nil {
					errored = istatus != 0
				}

				if errored {
					if res.GetMessage() == nil {
						return resp, fmt.Errorf("sdkerr: api error: unknown error")
					} else {
						return resp, fmt.Errorf("sdkerr: api error: message='%s'", *res.GetMessage())
					}
				}
			}
		}
	}

	return resp, nil
}

func generateSignature(timestamp string, apiKey string) string {
	keyMd5 := md5.Sum([]byte(apiKey))
	keyMd5Hex := strings.ToLower(hex.EncodeToString(keyMd5[:]))

	signMd5 := md5.Sum([]byte(timestamp + keyMd5Hex))
	signMd5Hex := strings.ToLower(hex.EncodeToString(signMd5[:]))
	return signMd5Hex
}
