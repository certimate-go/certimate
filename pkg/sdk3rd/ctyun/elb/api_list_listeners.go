package elb

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ListListenersRequest struct {
	ClientToken     *string `json:"clientToken,omitempty"     url:"clientToken,omitempty"`
	RegionID        *string `json:"regionID,omitempty"        url:"regionID,omitempty"`
	ProjectID       *string `json:"projectID,omitempty"       url:"projectID,omitempty"`
	IDs             *string `json:"IDs,omitempty"             url:"IDs,omitempty"`
	Name            *string `json:"name,omitempty"            url:"name,omitempty"`
	LoadBalancerID  *string `json:"loadBalancerID,omitempty"  url:"loadBalancerID,omitempty"`
	AccessControlID *string `json:"accessControlID,omitempty" url:"accessControlID,omitempty"`
}

type ListListenersResponse struct {
	sdkResponseBase

	ReturnObj []*Listener `json:"returnObj,omitempty"`
}

func (c *Client) ListListeners(req *ListListenersRequest) (*ListListenersResponse, error) {
	return c.ListListenersWithContext(context.Background(), req)
}

func (c *Client) ListListenersWithContext(ctx context.Context, req *ListListenersRequest) (*ListListenersResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v4/elb/list-listener")
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

	result := &ListListenersResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
