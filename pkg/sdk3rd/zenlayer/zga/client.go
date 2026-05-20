package zga

import (
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zga20230706 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

const (
	APIVersion = zga20230706.APIVersion
	SERVICE    = zga20230706.SERVICE
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

func NewCreateCertificateRequest() (request *CreateCertificateRequest) {
	return zga20230706.NewCreateCertificateRequest()
}

func NewCreateCertificateResponse() (response *CreateCertificateResponse) {
	return zga20230706.NewCreateCertificateResponse()
}

func (c *Client) CreateCertificate(request *CreateCertificateRequest) (response *CreateCertificateResponse, err error) {
	response = NewCreateCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func NewDescribeAcceleratorsRequest() (request *DescribeAcceleratorsRequest) {
	return zga20230706.NewDescribeAcceleratorsRequest()
}

func NewDescribeAcceleratorsResponse() (response *DescribeAcceleratorsResponse) {
	return zga20230706.NewDescribeAcceleratorsResponse()
}

func (c *Client) DescribeAccelerators(request *DescribeAcceleratorsRequest) (response *DescribeAcceleratorsResponse, err error) {
	response = NewDescribeAcceleratorsResponse()
	err = c.ApiCall(request, response)
	return
}

func NewDescribeCertificatesRequest() (request *DescribeCertificatesRequest) {
	return zga20230706.NewDescribeCertificatesRequest()
}

func NewDescribeCertificatesResponse() (response *DescribeCertificatesResponse) {
	return zga20230706.NewDescribeCertificatesResponse()
}

func (c *Client) DescribeCertificates(request *DescribeCertificatesRequest) (response *DescribeCertificatesResponse, err error) {
	response = NewDescribeCertificatesResponse()
	err = c.ApiCall(request, response)
	return
}

func NewModifyAcceleratorCertificateRequest() (request *ModifyAcceleratorCertificateRequest) {
	return zga20230706.NewModifyAcceleratorCertificateRequest()
}

func NewModifyAcceleratorCertificateResponse() (response *ModifyAcceleratorCertificateResponse) {
	return zga20230706.NewModifyAcceleratorCertificateResponse()
}

func (c *Client) ModifyAcceleratorCertificate(request *ModifyAcceleratorCertificateRequest) (response *ModifyAcceleratorCertificateResponse, err error) {
	response = NewModifyAcceleratorCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func NewModifyCertificateRequest() (request *ModifyCertificateRequest) {
	return zga20230706.NewModifyCertificateRequest()
}

func NewModifyCertificateResponse() (response *ModifyCertificateResponse) {
	return zga20230706.NewModifyCertificateResponse()
}

func (c *Client) ModifyCertificate(request *ModifyCertificateRequest) (response *ModifyCertificateResponse, err error) {
	response = NewModifyCertificateResponse()
	err = c.ApiCall(request, response)
	return
}
