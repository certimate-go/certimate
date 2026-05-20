package client

import (
	"encoding/json"
	"errors"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	cdn "github.com/jdcloud-api/jdcloud-sdk-go/services/cdn/apis"
)

type CdnClient struct {
	core.JDCloudClient
}

func NewCdnClient(credential *core.Credential) *CdnClient {
	if credential == nil {
		return nil
	}

	config := core.NewConfig()
	config.SetEndpoint("cdn.jdcloud-api.com")

	return &CdnClient{
		core.JDCloudClient{
			Credential:  *credential,
			Config:      *config,
			ServiceName: "cdn",
			Revision:    "0.10.48",
			Logger:      core.NewDefaultLogger(core.LogInfo),
		},
	}
}

func (c *CdnClient) DisableLogger() {
	c.Logger = core.NewDummyLogger()
}

func (c *CdnClient) GetDomainList(request *cdn.GetDomainListRequest) (*cdn.GetDomainListResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &cdn.GetDomainListResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *CdnClient) QueryDomainConfig(request *cdn.QueryDomainConfigRequest) (*cdn.QueryDomainConfigResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &cdn.QueryDomainConfigResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *CdnClient) SetHttpType(request *cdn.SetHttpTypeRequest) (*cdn.SetHttpTypeResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &cdn.SetHttpTypeResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}
