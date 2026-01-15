package synologydsm

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type LogoutResponse struct {
	sdkResponseBase
}

func (c *Client) Logout() (*LogoutResponse, error) {
	if c.sid == "" {
		return nil, nil
	}

	params := url.Values{
		"api":     {"SYNO.API.Auth"},
		"version": {strconv.Itoa(c.apiVersion)},
		"method":  {"logout"},
		"_sid":    {c.sid},
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/webapi/%s?%s", c.apiPath, params.Encode()))
	if err != nil {
		return nil, err
	}

	result := &LogoutResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	c.synoTokenMtx.Lock()
	defer c.synoTokenMtx.Unlock()
	c.sid = ""
	c.synoToken = ""

	return result, nil
}
