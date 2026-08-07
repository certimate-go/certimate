package v3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type DnsCreateRecordRequest struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
	Data *string `json:"data,omitempty"`
	TTL  *int    `json:"ttl,omitempty"`
}

type DnsCreateRecordResponse struct {
	sdkResponseBase

	UUID       string `json:"uuid"`
	DomainUUID string `json:"domain_uuid"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	TTL        int    `json:"ttl"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (c *Client) DnsCreateRecord(domainId string, req *DnsCreateRecordRequest) (*DnsCreateRecordResponse, error) {
	return c.DnsCreateRecordWithContext(context.Background(), domainId, req)
}

func (c *Client) DnsCreateRecordWithContext(ctx context.Context, domainId string, req *DnsCreateRecordRequest) (*DnsCreateRecordResponse, error) {
	if domainId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset domainId")
	}

	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	path := dnsBaseURL + fmt.Sprintf("/v1/domains/%s/records", url.PathEscape(domainId))
	httpreq, err := c.newRequest(http.MethodPost, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &DnsCreateRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
