package internal

import (
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

// This is a partial copy of https://github.com/zenlayer/zenlayercloud-sdk-go/blob/main/zenlayercloud/zga20230706/client.go
// to lightweight the vendor packages in the built binary.
type ZgaClient struct {
	*common.Client
}

func NewZgaClient(config *common.Config, secretKeyId, secretKeyPassword string) (client *ZgaClient, err error) {
	client = &ZgaClient{}

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

func (c *ZgaClient) CreateCertificate(request *zga.CreateCertificateRequest) (response *zga.CreateCertificateResponse, err error) {
	response = zga.NewCreateCertificateResponse()
	err = c.ApiCall(request, response)
	return
}

func (c *ZgaClient) DescribeCertificates(request *zga.DescribeCertificatesRequest) (response *zga.DescribeCertificatesResponse, err error) {
	response = zga.NewDescribeCertificatesResponse()
	err = c.ApiCall(request, response)
	return
}

func (c *ZgaClient) ModifyCertificate(request *zga.ModifyCertificateRequest) (response *zga.ModifyCertificateResponse, err error) {
	response = zga.NewModifyCertificateResponse()
	err = c.ApiCall(request, response)
	return
}
