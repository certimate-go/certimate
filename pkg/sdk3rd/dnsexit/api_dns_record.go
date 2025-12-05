package dnsexit

import (
	"context"
	"net/http"
)

type DnsRecordRequest struct {
	Domain *string    `json:"domain,omitempty"`
	Add    *DnsRecord `json:"add,omitempty"`
	Update *DnsRecord `json:"update,omitempty"`
	Delete *DnsRecord `json:"delete,omitempty"`
}

type DnsRecordResponse struct {
	apiResponseBase

	Details []string `json:"details,omitempty"`
}

func (c *Client) DnsRecord(req *DnsRecordRequest) (*DnsRecordResponse, error) {
	return c.DnsRecordWithContext(context.Background(), req)
}

func (c *Client) DnsRecordWithContext(ctx context.Context, req *DnsRecordRequest) (*DnsRecordResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/dns/")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &DnsRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
