package internal

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwaadv1 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1"
	modelv1 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1/model"
	hwaadv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v2"
	modelv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v2/model"
)

// This is a partial copy of https://github.com/huaweicloud/huaweicloud-sdk-go-v3/blob/master/services/aad/v2/aad_client.go
// to lightweight the vendor packages in the built binary.
type AadClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewAadClient(hcClient *httpclient.HcHttpClient) *AadClient {
	return &AadClient{HcClient: hcClient}
}

func (c *AadClient) ListInstanceDomains(request *modelv2.ListInstanceDomainsRequest) (*modelv2.ListInstanceDomainsResponse, error) {
	requestDef := hwaadv2.GenReqDefForListInstanceDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*modelv2.ListInstanceDomainsResponse), nil
	}
}

func (c *AadClient) SetCertForDomain(request *modelv1.SetCertForDomainRequest) (*modelv1.SetCertForDomainResponse, error) {
	requestDef := hwaadv1.GenReqDefForSetCertForDomain()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*modelv1.SetCertForDomainResponse), nil
	}
}
