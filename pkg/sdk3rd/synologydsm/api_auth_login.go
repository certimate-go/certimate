package synologydsm

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	qs "github.com/google/go-querystring/query"
)

type LoginRequest struct {
	Account  string `json:"account"            url:"account"`
	Password string `json:"passwd"             url:"passwd"`
	OtpCode  string `json:"otp_code,omitempty" url:"otp_code,omitempty"`
}

type LoginResponse struct {
	sdkResponseBase
	Data *struct {
		Sid       string `json:"sid"`
		SynoToken string `json:"synotoken"`
		DeviceId  string `json:"device_id,omitempty"`
		Did       string `json:"did,omitempty"`
	} `json:"data,omitempty"`
}

func (c *Client) Login(req *LoginRequest) (*LoginResponse, error) {
	if c.apiPath == "" || c.apiVersion == 0 {
		queryInfoReq := &QueryAPIInfoRequest{
			Query: "SYNO.API.Auth",
		}
		queryInfoResp, err := c.QueryAPIInfo(queryInfoReq)
		if err != nil {
			return nil, fmt.Errorf("sdkerr: failed to query API info: %w", err)
		} else {
			authInfo, ok := queryInfoResp.Data["SYNO.API.Auth"]
			if !ok {
				return nil, fmt.Errorf("sdkerr: failed to query API info: \"SYNO.API.Auth\" not found")
			}

			c.apiPath = authInfo.Path
			c.apiVersion = authInfo.MaxVersion
		}
	}

	params := url.Values{
		"api":                 {"SYNO.API.Auth"},
		"version":             {strconv.Itoa(c.apiVersion)},
		"method":              {"login"},
		"format":              {"sid"},
		"enable_syno_token":   {"yes"},
		"enable_device_token": {"yes"},
		"device_name":         {"Certimate"},
	}

	values, err := qs.Values(req)
	if err != nil {
		return nil, err
	}
	for k := range values {
		params.Set(k, values.Get(k))
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/webapi/%s?%s", c.apiPath, params.Encode()))
	if err != nil {
		return nil, err
	}

	result := &LoginResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		if result != nil && result.GetErrorCode() > 0 {
			errcode := result.GetErrorCode()
			errdesc := getAuthErrorDescription(errcode)
			return result, fmt.Errorf("sdkerr: code='%d', desc='%s'", errcode, errdesc)
		}
		return result, err
	}

	if result.Data.Sid == "" || result.Data.SynoToken == "" {
		return result, fmt.Errorf("sdkerr: login succeeded but the sid or synotoken is empty")
	}

	c.synoTokenMtx.Lock()
	defer c.synoTokenMtx.Unlock()
	c.sid = result.Data.Sid
	c.synoToken = result.Data.SynoToken

	return result, nil
}
