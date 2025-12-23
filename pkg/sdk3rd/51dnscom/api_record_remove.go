package dnscom

import (
	"context"
	"net/http"
)

type RecordRemoveRequest struct {
	DomainID *string `json:"domainID,omitempty"`
	RecordID *string `json:"recordID,omitempty"`
}

type RecordRemoveResponse struct {
	apiResponseBase
}

func (c *Client) RecordRemove(req *RecordRemoveRequest) (*RecordRemoveResponse, error) {
	return c.RecordRemoveWithContext(context.Background(), req)
}

func (c *Client) RecordRemoveWithContext(ctx context.Context, req *RecordRemoveRequest) (*RecordRemoveResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/record/remove/", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &RecordRemoveResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
