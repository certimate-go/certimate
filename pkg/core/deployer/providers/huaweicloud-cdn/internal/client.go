package internal

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwcdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/model"
)

// This is a partial copy of https://github.com/huaweicloud/huaweicloud-sdk-go-v3/blob/master/services/cdn/v2/cdn_client.go
// to lightweight the vendor packages in the built binary.
type CdnClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewCdnClient(hcClient *httpclient.HcHttpClient) *CdnClient {
	return &CdnClient{HcClient: hcClient}
}

func (c *CdnClient) ListDomains(request *model.ListDomainsRequest) (*model.ListDomainsResponse, error) {
	requestDef := hwcdn.GenReqDefForListDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListDomainsResponse), nil
	}
}

func (c *CdnClient) UpdateDomainMultiCertificates(request *model.UpdateDomainMultiCertificatesRequest) (*model.UpdateDomainMultiCertificatesResponse, error) {
	requestDef := hwcdn.GenReqDefForUpdateDomainMultiCertificates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainMultiCertificatesResponse), nil
	}
}
