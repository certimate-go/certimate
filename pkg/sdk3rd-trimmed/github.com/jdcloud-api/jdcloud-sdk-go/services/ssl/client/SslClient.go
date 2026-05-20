package client

import (
	"encoding/json"
	"errors"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	ssl "github.com/jdcloud-api/jdcloud-sdk-go/services/ssl/apis"
)

type SslClient struct {
	core.JDCloudClient
}

func NewSslClient(credential *core.Credential) *SslClient {
	if credential == nil {
		return nil
	}

	config := core.NewConfig()
	config.SetEndpoint("ssl.jdcloud-api.com")

	return &SslClient{
		core.JDCloudClient{
			Credential:  *credential,
			Config:      *config,
			ServiceName: "ssl",
			Revision:    "1.0.2",
			Logger:      core.NewDefaultLogger(core.LogInfo),
		},
	}
}

func (c *SslClient) DisableLogger() {
	c.Logger = core.NewDummyLogger()
}

func (c *SslClient) DescribeCerts(request *ssl.DescribeCertsRequest) (*ssl.DescribeCertsResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &ssl.DescribeCertsResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *SslClient) UploadCert(request *ssl.UploadCertRequest) (*ssl.UploadCertResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &ssl.UploadCertResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}
