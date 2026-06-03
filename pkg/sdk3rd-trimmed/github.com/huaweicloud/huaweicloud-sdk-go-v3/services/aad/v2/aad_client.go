package v2

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	aad "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v2/model"
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

func (c *AadClient) ListInstanceDomains(request *model.ListInstanceDomainsRequest) (*model.ListInstanceDomainsResponse, error) {
	requestDef := GenReqDefForListInstanceDomains()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListInstanceDomainsResponse), nil
	}
}
