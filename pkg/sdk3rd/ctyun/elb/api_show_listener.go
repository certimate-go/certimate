package elb

import (
	"context"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type ShowListenerRequest struct {
	ClientToken *string `json:"clientToken,omitempty" url:"clientToken,omitempty"`
	RegionID    *string `json:"regionID,omitempty"    url:"regionID,omitempty"`
	ListenerID  *string `json:"listenerID,omitempty"  url:"listenerID,omitempty"`
}

type ShowListenerResponse struct {
	sdkResponseBase

	ReturnObj []*Listener `json:"returnObj,omitempty"`
}

func (c *Client) ShowListener(req *ShowListenerRequest) (*ShowListenerResponse, error) {
	return c.ShowListenerWithContext(context.Background(), req)
}

func (c *Client) ShowListenerWithContext(ctx context.Context, req *ShowListenerRequest) (*ShowListenerResponse, error) {
	httpreq, err := c.newRequest(http.MethodGet, "/v4/elb/show-listener")
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

	result := &ShowListenerResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
