package client

import (
	"encoding/json"
	"errors"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	vod "github.com/jdcloud-api/jdcloud-sdk-go/services/vod/apis"
)

type VodClient struct {
	core.JDCloudClient
}

func NewVodClient(credential *core.Credential) *VodClient {
	if credential == nil {
		return nil
	}

	config := core.NewConfig()
	config.SetEndpoint("vod.jdcloud-api.com")

	return &VodClient{
		core.JDCloudClient{
			Credential:  *credential,
			Config:      *config,
			ServiceName: "vod",
			Revision:    "1.2.1",
			Logger:      core.NewDefaultLogger(core.LogInfo),
		},
	}
}

func (c *VodClient) DisableLogger() {
	c.Logger = core.NewDummyLogger()
}

func (c *VodClient) ListDomains(request *vod.ListDomainsRequest) (*vod.ListDomainsResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &vod.ListDomainsResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *VodClient) GetHttpSsl(request *vod.GetHttpSslRequest) (*vod.GetHttpSslResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &vod.GetHttpSslResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *VodClient) SetHttpSsl(request *vod.SetHttpSslRequest) (*vod.SetHttpSslResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &vod.SetHttpSslResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}
