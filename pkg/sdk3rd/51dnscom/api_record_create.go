package dnscom

import (
	"context"
	"net/http"
)

type RecordCreateRequest struct {
	DomainID *int64  `json:"domainID,omitempty"`
	ViewID   *int64  `json:"viewID,omitempty"`
	Type     *string `json:"type,omitempty"`
	Host     *string `json:"host,omitempty"`
	Value    *string `json:"value,omitempty"`
	TTL      *int32  `json:"ttl,omitempty"`
	MX       *int32  `json:"mx,omitempty"`
	Remark   *string `json:"remark,omitempty"`
}

type RecordCreateResponse struct {
	apiResponseBase

	Data *DNSRecord `json:"data"`
}

func (c *Client) RecordCreate(req *RecordCreateRequest) (*RecordCreateResponse, error) {
	return c.RecordCreateWithContext(context.Background(), req)
}

func (c *Client) RecordCreateWithContext(ctx context.Context, req *RecordCreateRequest) (*RecordCreateResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/record/create/", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &RecordCreateResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
