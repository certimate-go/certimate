package synologydsm

import (
	"fmt"
	"net/http"
	"net/url"

	qs "github.com/google/go-querystring/query"
)

type QueryAPIInfoRequest struct {
	Query string `json:"query" url:"query"`
}

type QueryAPIInfoResponse struct {
	sdkResponseBase
	Data map[string]APIInfo `json:"data,omitempty"`
}

func (c *Client) QueryAPIInfo(req *QueryAPIInfoRequest) (*QueryAPIInfoResponse, error) {
	params := url.Values{
		"api":     {"SYNO.API.Info"},
		"version": {"1"},
		"method":  {"query"},
	}

	values, err := qs.Values(req)
	if err != nil {
		return nil, err
	}
	for k := range values {
		params.Set(k, values.Get(k))
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/webapi/query.cgi?%s", params.Encode()))
	if err != nil {
		return nil, err
	}

	result := &QueryAPIInfoResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
