package v2

import (
	"context"
	"net/http"
)

type WebsiteSSLSearchRequest struct {
	Domain   string `json:"domain"`
	Order    string `json:"order"`
	OrderBy  string `json:"orderBy"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"pageSize"`
}

type WebsiteSSLSearchResponse struct {
	sdkResponseBase

	Data *struct {
		Items []*struct {
			ID          int64  `json:"id"`
			PEM         string `json:"pem"`
			PrivateKey  string `json:"privateKey"`
			Domains     string `json:"domains"`
			Description string `json:"description"`
			Status      string `json:"status"`
			UpdatedAt   string `json:"updatedAt"`
			CreatedAt   string `json:"createdAt"`
		} `json:"items"`
		Total int32 `json:"total"`
	} `json:"data,omitempty"`
}

func (c *Client) WebsiteSSLSearch(req *WebsiteSSLSearchRequest) (*WebsiteSSLSearchResponse, error) {
	return c.WebsiteSSLSearchWithContext(context.Background(), req)
}

func (c *Client) WebsiteSSLSearchWithContext(ctx context.Context, req *WebsiteSSLSearchRequest) (*WebsiteSSLSearchResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/websites/ssl/search")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &WebsiteSSLSearchResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
