package client

import (
	"encoding/json"
	"errors"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	live "github.com/jdcloud-api/jdcloud-sdk-go/services/live/apis"
)

type LiveClient struct {
	core.JDCloudClient
}

func NewLiveClient(credential *core.Credential) *LiveClient {
	if credential == nil {
		return nil
	}

	config := core.NewConfig()
	config.SetEndpoint("live.jdcloud-api.com")

	return &LiveClient{
		core.JDCloudClient{
			Credential:  *credential,
			Config:      *config,
			ServiceName: "live",
			Revision:    "1.0.22",
			Logger:      core.NewDefaultLogger(core.LogInfo),
		},
	}
}

func (c *LiveClient) DisableLogger() {
	c.Logger = core.NewDummyLogger()
}

func (c *LiveClient) DescribeLiveDomains(request *live.DescribeLiveDomainsRequest) (*live.DescribeLiveDomainsResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &live.DescribeLiveDomainsResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *LiveClient) SetLiveDomainCertificate(request *live.SetLiveDomainCertificateRequest) (*live.SetLiveDomainCertificateResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &live.SetLiveDomainCertificateResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}
