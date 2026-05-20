package v3

import (
	httpclient "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	hwscm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/scm/v3/model"
)

type ScmClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewScmClient(hcClient *httpclient.HcHttpClient) *ScmClient {
	return &ScmClient{HcClient: hcClient}
}

func ScmClientBuilder() *httpclient.HcHttpClientBuilder {
	return hwscm.ScmClientBuilder()
}

func (c *ScmClient) ExportCertificate(request *model.ExportCertificateRequest) (*model.ExportCertificateResponse, error) {
	requestDef := GenReqDefForExportCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExportCertificateResponse), nil
	}
}

func (c *ScmClient) ImportCertificate(request *model.ImportCertificateRequest) (*model.ImportCertificateResponse, error) {
	requestDef := GenReqDefForImportCertificate()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ImportCertificateResponse), nil
	}
}

func (c *ScmClient) ListCertificates(request *model.ListCertificatesRequest) (*model.ListCertificatesResponse, error) {
	requestDef := GenReqDefForListCertificates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListCertificatesResponse), nil
	}
}
