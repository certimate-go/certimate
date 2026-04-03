package internal

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwapig "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
)

// This is a partial copy of https://github.com/huaweicloud/huaweicloud-sdk-go-v3/blob/master/services/apig/v2/apig_client.go
// to lightweight the vendor packages in the built binary.
type ApigClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewApigClient(hcClient *httpclient.HcHttpClient) *ApigClient {
	return &ApigClient{HcClient: hcClient}
}

func (c *ApigClient) ShowDetailsOfCertificateV2(request *model.ShowDetailsOfCertificateV2Request) (*model.ShowDetailsOfCertificateV2Response, error) {
	requestDef := hwapig.GenReqDefForShowDetailsOfCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowDetailsOfCertificateV2Response), nil
	}
}

func (c *ApigClient) UpdateCertificateV2(request *model.UpdateCertificateV2Request) (*model.UpdateCertificateV2Response, error) {
	requestDef := hwapig.GenReqDefForUpdateCertificateV2()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateCertificateV2Response), nil
	}
}
