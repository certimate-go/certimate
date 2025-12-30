package nginxproxymanager

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type NginxListDeadHostsRequest struct {
	Expand *string `json:"expand,omitempty" url:"expand,omitempty"`
}

type NginxListDeadHostsResponse = []*DeadHostRecord

func (c *Client) NginxListDeadHosts(req *NginxListDeadHostsRequest) (*NginxListDeadHostsResponse, error) {
	return c.NginxListDeadHostsWithContext(context.Background(), req)
}

func (c *Client) NginxListDeadHostsWithContext(ctx context.Context, req *NginxListDeadHostsRequest) (*NginxListDeadHostsResponse, error) {
	if err := c.ensureJwtTokenExists(); err != nil {
		return nil, err
	}

	httpreq, err := c.newRequest(http.MethodGet, "/nginx/dead-hosts")
	if err != nil {
		return nil, err
	} else {
		values, err := qs.Values(req)
		if err != nil {
			return nil, err
		}

		httpreq.SetQueryParamsFromValues(values)
		httpreq.SetContext(ctx)
	}

	result := &NginxListDeadHostsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
