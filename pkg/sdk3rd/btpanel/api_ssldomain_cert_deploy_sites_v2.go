package btpanel

import (
	"context"
	"net/http"
)

type SSLDomainCertDeploySitesV2Request struct {
	SSLHash string   `json:"hash"`
	Domains []string `json:"domains"`
	Append  int32    `json:"append"`
}

type SSLDomainCertDeploySitesV2Response struct {
	sdkResponseBaseV2
}

func (c *Client) SSLDomainCertDeploySitesV2(req *SSLDomainCertDeploySitesV2Request) (*SSLDomainCertDeploySitesV2Response, error) {
	return c.SSLDomainCertDeploySitesV2WithContext(context.Background(), req)
}

func (c *Client) SSLDomainCertDeploySitesV2WithContext(ctx context.Context, req *SSLDomainCertDeploySitesV2Request) (*SSLDomainCertDeploySitesV2Response, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/v2/ssl_domain?action=cert_deploy_sites", req)
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &SSLDomainCertDeploySitesV2Response{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
