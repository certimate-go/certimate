package dynadot

import (
	"context"
	"fmt"
	"net/http"
)

type SetDnsRequest struct {
	DnsMainList            []*DnsMainRecord `json:"dns_main_list,omitempty"`
	SubList                []*DnsSubRecord  `json:"sub_list,omitempty"`
	TTL                    *int64           `json:"ttl,omitempty"`
	AddDnsToCurrentSetting *bool            `json:"add_dns_to_current_setting,omitempty"`
}

type SetDnsResponse struct {
	sdkResponseBase
}

func (c *Client) SetDns(domain string, req *SetDnsRequest) (*SetDnsResponse, error) {
	return c.SetDnsWithContext(context.Background(), domain, req)
}

func (c *Client) SetDnsWithContext(ctx context.Context, domain string, req *SetDnsRequest) (*SetDnsResponse, error) {
	if domain == "" {
		return nil, fmt.Errorf("sdkerr: unset domain")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/restful/v2/domains/%s/records", domain), req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SetDnsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
