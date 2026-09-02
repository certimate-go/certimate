package asiaispcdn

import (
	"context"
	"net/http"
)

type DomainModifyRequest struct {
	Domain   *string `json:"domain,omitempty"`
	Protocol *string `json:"protocol,omitempty"`
	CertId   *int64  `json:"certId,omitempty"`
}

type DomainModifyResponse struct {
	sdkResponseBase

	Data int64 `json:"data,omitempty"`
}

func (c *Client) DomainModify(req *DomainModifyRequest) (*DomainModifyResponse, error) {
	return c.DomainModifyWithContext(context.Background(), req)
}

func (c *Client) DomainModifyWithContext(ctx context.Context, req *DomainModifyRequest) (*DomainModifyResponse, error) {
	httpreq, err := c.newRequest(http.MethodPut, req, "domainModify")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DomainModifyResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
