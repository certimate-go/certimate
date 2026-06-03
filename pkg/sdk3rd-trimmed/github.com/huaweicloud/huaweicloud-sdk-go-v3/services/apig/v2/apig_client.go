package v2

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	apig "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
)

type ApigClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewApigClient(hcClient *httpclient.HcHttpClient) *ApigClient {
	return &ApigClient{HcClient: hcClient}
}

func ApigClientBuilder() *httpclient.HcHttpClientBuilder {
	return apig.ApigClientBuilder()
}

func (c *ApigClient) ShowDetailsOfCertificateV2(request *model.ShowDetailsOfCertificateV2Request) (*model.ShowDetailsOfCertificateV2Response, error) {
	requestDef := GenReqDefForShowDetailsOfCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfCertificateV2Response), nil
	}
}

func (c *ApigClient) UpdateCertificateV2(request *model.UpdateCertificateV2Request) (*model.UpdateCertificateV2Response, error) {
	requestDef := GenReqDefForUpdateCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateCertificateV2Response), nil
	}
}
