package v2

import (
	"context"
	"fmt"
	"net/http"
)

type WebsiteHttpsPostRequest struct {
	WebsiteID    int64    `json:"websiteId"`
	Enable       bool     `json:"enable"`
	Type         string   `json:"type"`
	WebsiteSSLID int64    `json:"websiteSSLId"`
	HttpConfig   string   `json:"httpConfig"`
	SSLProtocol  []string `json:"SSLProtocol"`
	Algorithm    string   `json:"algorithm"`
	Hsts         bool     `json:"hsts"`
	Http3        bool     `json:"http3"`
}

type WebsiteHttpsPostResponse struct {
	sdkResponseBase
}

func (c *Client) WebsiteHttpsPost(websiteId int64, req *WebsiteHttpsPostRequest) (*WebsiteHttpsPostResponse, error) {
	return c.WebsiteHttpsPostWithContext(context.Background(), websiteId, req)
}

func (c *Client) WebsiteHttpsPostWithContext(ctx context.Context, websiteId int64, req *WebsiteHttpsPostRequest) (*WebsiteHttpsPostResponse, error) {
	if websiteId == 0 {
		return nil, fmt.Errorf("sdkerr: unset websiteId")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/websites/%d/https", websiteId))
	if err != nil {
		return nil, err
	} else {
		req.WebsiteID = websiteId
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &WebsiteHttpsPostResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
