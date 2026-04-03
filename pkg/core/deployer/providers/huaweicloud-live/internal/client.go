package internal

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwlive "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/live/v1/model"
)

// This is a partial copy of https://github.com/huaweicloud/huaweicloud-sdk-go-v3/blob/master/services/live/v1/live_client.go
// to lightweight the vendor packages in the built binary.
type LiveClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewLiveClient(hcClient *httpclient.HcHttpClient) *LiveClient {
	return &LiveClient{HcClient: hcClient}
}

func (c *LiveClient) ShowDomain(request *model.ShowDomainRequest) (*model.ShowDomainResponse, error) {
	requestDef := hwlive.GenReqDefForShowDomain()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDomainResponse), nil
	}
}

func (c *LiveClient) UpdateDomainHttpsCert(request *model.UpdateDomainHttpsCertRequest) (*model.UpdateDomainHttpsCertResponse, error) {
	requestDef := hwlive.GenReqDefForUpdateDomainHttpsCert()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateDomainHttpsCertResponse), nil
	}
}
