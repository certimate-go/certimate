package dynadot

import (
	"context"
	"fmt"
	"net/http"
)

type RemoveDnsRequest struct {
	DnsMainList []*DnsMainRecord `json:"dns_main_list,omitempty"`
	SubList     []*DnsSubRecord  `json:"sub_list,omitempty"`
}

type RemoveDnsResponse struct {
	sdkResponseBase
}

func (c *Client) RemoveDns(domain string, req *RemoveDnsRequest) (*RemoveDnsResponse, error) {
	return c.RemoveDnsWithContext(context.Background(), domain, req)
}

func (c *Client) RemoveDnsWithContext(ctx context.Context, domain string, req *RemoveDnsRequest) (*RemoveDnsResponse, error) {
	if domain == "" {
		return nil, fmt.Errorf("sdkerr: unset domain")
	}

	httpreq, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/restful/v2/domains/%s/records", domain), req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &RemoveDnsResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
