package v1

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	waf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"
)

type WafClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewWafClient(hcClient *httpclient.HcHttpClient) *WafClient {
	return &WafClient{HcClient: hcClient}
}

func WafClientBuilder() *httpclient.HcHttpClientBuilder {
	return waf.WafClientBuilder()
}

func (c *WafClient) CreateCertificate(request *model.CreateCertificateRequest) (*model.CreateCertificateResponse, error) {
	requestDef := GenReqDefForCreateCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateCertificateResponse), nil
	}
}

func (c *WafClient) ListCertificates(request *model.ListCertificatesRequest) (*model.ListCertificatesResponse, error) {
	requestDef := GenReqDefForListCertificates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCertificatesResponse), nil
	}
}

func (c *WafClient) ListHost(request *model.ListHostRequest) (*model.ListHostResponse, error) {
	requestDef := GenReqDefForListHost()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListHostResponse), nil
	}
}

func (c *WafClient) ListPremiumHost(request *model.ListPremiumHostRequest) (*model.ListPremiumHostResponse, error) {
	requestDef := GenReqDefForListPremiumHost()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPremiumHostResponse), nil
	}
}

func (c *WafClient) ShowCertificate(request *model.ShowCertificateRequest) (*model.ShowCertificateResponse, error) {
	requestDef := GenReqDefForShowCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowCertificateResponse), nil
	}
}

func (c *WafClient) UpdateCertificate(request *model.UpdateCertificateRequest) (*model.UpdateCertificateResponse, error) {
	requestDef := GenReqDefForUpdateCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateCertificateResponse), nil
	}
}

func (c *WafClient) UpdateHost(request *model.UpdateHostRequest) (*model.UpdateHostResponse, error) {
	requestDef := GenReqDefForUpdateHost()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateHostResponse), nil
	}
}

func (c *WafClient) UpdatePremiumHost(request *model.UpdatePremiumHostRequest) (*model.UpdatePremiumHostResponse, error) {
	requestDef := GenReqDefForUpdatePremiumHost()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdatePremiumHostResponse), nil
	}
}
