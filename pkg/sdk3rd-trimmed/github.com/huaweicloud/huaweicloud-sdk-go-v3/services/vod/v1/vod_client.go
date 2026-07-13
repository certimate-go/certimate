package v1

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	vod "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vod/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vod/v1/model"
)

type VodClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewVodClient(hcClient *httpclient.HcHttpClient) *VodClient {
	return &VodClient{HcClient: hcClient}
}

func VodClientBuilder() *httpclient.HcHttpClientBuilder {
	return vod.VodClientBuilder()
}

func (c *VodClient) ShowHttpsConfig(request *model.ShowHttpsConfigRequest) (*model.ShowHttpsConfigResponse, error) {
	requestDef := GenReqDefForShowHttpsConfig()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowHttpsConfigResponse), nil
	}
}

func (c *VodClient) UpdateHttpsConfig(request *model.UpdateHttpsConfigRequest) (*model.UpdateHttpsConfigResponse, error) {
	requestDef := GenReqDefForUpdateHttpsConfig()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateHttpsConfigResponse), nil
	}
}
