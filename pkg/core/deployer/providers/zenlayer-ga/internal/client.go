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

func (c *ZgaClient) DescribeAccelerators(request *zga.DescribeAcceleratorsRequest) (response *zga.DescribeAcceleratorsResponse, err error) {
	response = zga.NewDescribeAcceleratorsResponse()
	err = c.ApiCall(request, response)
	return
}

func (c *ZgaClient) ModifyAcceleratorCertificate(request *zga.ModifyAcceleratorCertificateRequest) (response *zga.ModifyAcceleratorCertificateResponse, err error) {
	response = zga.NewModifyAcceleratorCertificateResponse()
	err = c.ApiCall(request, response)
	return
}
