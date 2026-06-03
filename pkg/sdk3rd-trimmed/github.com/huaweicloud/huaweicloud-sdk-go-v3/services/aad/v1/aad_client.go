package v1

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	aad "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1/model"
)

type AadClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewAadClient(hcClient *httpclient.HcHttpClient) *AadClient {
	return &AadClient{HcClient: hcClient}
}

func AadClientBuilder() *httpclient.HcHttpClientBuilder {
	return aad.AadClientBuilder()
}

func (c *AadClient) SetCertForDomain(request *model.SetCertForDomainRequest) (*model.SetCertForDomainResponse, error) {
	requestDef := GenReqDefForSetCertForDomain()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.SetCertForDomainResponse), nil
	}
}
