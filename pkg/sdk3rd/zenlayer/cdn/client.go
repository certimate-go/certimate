package cdn

import (
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

const (
	APIVersion = "2022-11-20"
	SERVICE    = "cdn"
)

type Client struct {
	common.Client
}

func NewClientWithSecretKey(secretKeyId, secretKeyPassword string) (client *Client, err error) {
	return NewClient(common.NewConfig(), secretKeyId, secretKeyPassword)
}

func NewClient(config *common.Config, secretKeyId, secretKeyPassword string) (client *Client, err error) {
	client = &Client{}

	err = client.InitWithCredential(common.NewCredential(secretKeyId, secretKeyPassword))
	if err != nil {
		return nil, err
	}
	err = client.WithConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewDescribeCertificatesRequest() (request *DescribeCertificatesRequest) {
	request = &DescribeCertificatesRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "DescribeCertificates")

	return
}

func NewDescribeCertificatesResponse() (response *DescribeCertificatesResponse) {
	response = &DescribeCertificatesResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) DescribeCertificates(request *DescribeCertificatesRequest) (response *DescribeCertificatesResponse, err error) {
	response = NewDescribeCertificatesResponse()
	err = c.ApiCall(request, response)
	return
}

func NewCreateCertificateRequest() (request *CreateCertificateRequest) {
	request = &CreateCertificateRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "CreateCertificate")

	return
}

func NewCreateCertificateResponse() (response *CreateCertificateResponse) {
	response = &CreateCertificateResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) CreateCertificate(request *CreateCertificateRequest) (response *CreateCertificateResponse, err error) {
	response = NewCreateCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func NewModifyCertificateRequest() (request *ModifyCertificateRequest) {
	request = &ModifyCertificateRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "ModifyCertificate")

	return
}

func NewModifyCertificateResponse() (response *ModifyCertificateResponse) {
	response = &ModifyCertificateResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) ModifyCertificate(request *ModifyCertificateRequest) (response *ModifyCertificateResponse, err error) {
	response = NewModifyCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func NewDeleteCertificateRequest() (request *DeleteCertificateRequest) {
	request = &DeleteCertificateRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "DeleteCertificate")

	return
}

func NewDeleteCertificateResponse() (response *DeleteCertificateResponse) {
	response = &DeleteCertificateResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) DeleteCertificate(request *DeleteCertificateRequest) (response *DeleteCertificateResponse, err error) {
	response = NewDeleteCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func NewDescribeDomainsRequest() (request *DescribeDomainsRequest) {
	request = &DescribeDomainsRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "DescribeDomains")

	return
}

func NewDescribeDomainsResponse() (response *DescribeDomainsResponse) {
	response = &DescribeDomainsResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) DescribeDomains(request *DescribeDomainsRequest) (response *DescribeDomainsResponse, err error) {
	response = NewDescribeDomainsResponse()
	err = c.ApiCall(request, response)
	return
}

func NewDescribeDomainCertificateRequest() (request *DescribeDomainCertificateRequest) {
	request = &DescribeDomainCertificateRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "DescribeDomainCertificate")

	return
}

func NewDescribeDomainCertificateResponse() (response *DescribeDomainCertificateResponse) {
	response = &DescribeDomainCertificateResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) DescribeDomainCertificate(request *DescribeDomainCertificateRequest) (response *DescribeDomainCertificateResponse, err error) {
	response = NewDescribeDomainCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func NewModifyDomainCertificateRequest() (request *ModifyDomainCertificateRequest) {
	request = &ModifyDomainCertificateRequest{
		BaseRequest: &common.BaseRequest{},
	}
	request.Init().InitWithApiInfo(SERVICE, APIVersion, "ModifyDomainCertificate")

	return
}

func NewModifyDomainCertificateResponse() (response *ModifyDomainCertificateResponse) {
	response = &ModifyDomainCertificateResponse{
		BaseResponse: &common.BaseResponse{},
	}
	return
}

func (c *Client) ModifyDomainCertificate(request *ModifyDomainCertificateRequest) (response *ModifyDomainCertificateResponse, err error) {
	response = NewModifyDomainCertificateResponse()
	err = c.ApiCall(request, response)
	return
}
