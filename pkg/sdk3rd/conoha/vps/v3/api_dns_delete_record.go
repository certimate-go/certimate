package v3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type DnsDeleteRecordResponse struct {
	sdkResponseBase
}

func (c *Client) DnsDeleteRecord(domainId string, recordId string) (*DnsDeleteRecordResponse, error) {
	return c.DnsDeleteRecordWithContext(context.Background(), domainId, recordId)
}

func (c *Client) DnsDeleteRecordWithContext(ctx context.Context, domainId string, recordId string) (*DnsDeleteRecordResponse, error) {
	if domainId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset domainId")
	}
	if recordId == "" {
		return nil, fmt.Errorf("sdkerr: bad request: unset recordId")
	}

	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	path := dnsBaseURL + fmt.Sprintf("/v1/domains/%s/records/%s", url.PathEscape(domainId), url.PathEscape(recordId))
	httpreq, err := c.newRequest(http.MethodDelete, path)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &DnsDeleteRecordResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
