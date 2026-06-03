package v1

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	live "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v1/model"
)

type LiveClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewLiveClient(hcClient *httpclient.HcHttpClient) *LiveClient {
	return &LiveClient{HcClient: hcClient}
}

func LiveClientBuilder() *httpclient.HcHttpClientBuilder {
	return live.LiveClientBuilder()
}

func (c *LiveClient) ShowDomain(request *model.ShowDomainRequest) (*model.ShowDomainResponse, error) {
	requestDef := GenReqDefForShowDomain()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainResponse), nil
	}
}

func (c *LiveClient) UpdateDomainHttpsCert(request *model.UpdateDomainHttpsCertRequest) (*model.UpdateDomainHttpsCertResponse, error) {
	requestDef := GenReqDefForUpdateDomainHttpsCert()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainHttpsCertResponse), nil
	}
}
