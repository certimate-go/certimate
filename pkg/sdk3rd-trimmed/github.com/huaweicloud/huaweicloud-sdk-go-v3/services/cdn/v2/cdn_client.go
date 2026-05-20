package v2

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwcdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/model"
)

type CdnClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewCdnClient(hcClient *httpclient.HcHttpClient) *CdnClient {
	return &CdnClient{HcClient: hcClient}
}

func CdnClientBuilder() *httpclient.HcHttpClientBuilder {
	return hwcdn.CdnClientBuilder()
}

func (c *CdnClient) ListDomains(request *model.ListDomainsRequest) (*model.ListDomainsResponse, error) {
	requestDef := GenReqDefForListDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainsResponse), nil
	}
}

func (c *CdnClient) UpdateDomainMultiCertificates(request *model.UpdateDomainMultiCertificatesRequest) (*model.UpdateDomainMultiCertificatesResponse, error) {
	requestDef := GenReqDefForUpdateDomainMultiCertificates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainMultiCertificatesResponse), nil
	}
}
