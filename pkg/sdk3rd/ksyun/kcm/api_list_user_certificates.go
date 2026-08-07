package kcm

import (
	"context"
	"net/http"
)

type ListUserCertificatesRequest struct {
	Page     *int32 `json:"Page,omitempty"`
	PageSize *int32 `json:"PageSize,omitempty"`
}

type ListUserCertificatesResponse struct {
	sdkResponseBase

	Success bool `json:"Success"`
	Ret     *struct {
		Certs []*UserCertificate `json:"Certs"`
	} `json:"Ret,omitempty"`
}

func (c *Client) ListUserCertificates(req *ListUserCertificatesRequest) (*ListUserCertificatesResponse, error) {
	return c.ListUserCertificatesWithContext(context.Background(), req)
}

func (c *Client) ListUserCertificatesWithContext(ctx context.Context, req *ListUserCertificatesRequest) (*ListUserCertificatesResponse, error) {
	params := &struct {
		ListUserCertificatesRequest `json:",inline"`
		Action                      string
		Version                     string
	}{
		ListUserCertificatesRequest: *req,
		Action:                      "ListUserCertificates",
		Version:                     "2016-03-04",
	}

	httpreq, err := c.newRequest(http.MethodGet, "/", params)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &ListUserCertificatesResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
