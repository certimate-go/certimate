package client

import (
	"encoding/json"
	"errors"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	lb "github.com/jdcloud-api/jdcloud-sdk-go/services/lb/apis"
)

type LbClient struct {
	core.JDCloudClient
}

func NewLbClient(credential *core.Credential) *LbClient {
	if credential == nil {
		return nil
	}

	config := core.NewConfig()
	config.SetEndpoint("lb.jdcloud-api.com")

	return &LbClient{
		core.JDCloudClient{
			Credential:  *credential,
			Config:      *config,
			ServiceName: "lb",
			Revision:    "0.6.6",
			Logger:      core.NewDefaultLogger(core.LogInfo),
		},
	}
}

func (c *LbClient) DisableLogger() {
	c.Logger = core.NewDummyLogger()
}

func (c *LbClient) DescribeListener(request *lb.DescribeListenerRequest) (*lb.DescribeListenerResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &lb.DescribeListenerResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *LbClient) DescribeListeners(request *lb.DescribeListenersRequest) (*lb.DescribeListenersResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &lb.DescribeListenersResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *LbClient) DescribeLoadBalancer(request *lb.DescribeLoadBalancerRequest) (*lb.DescribeLoadBalancerResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &lb.DescribeLoadBalancerResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *LbClient) UpdateListener(request *lb.UpdateListenerRequest) (*lb.UpdateListenerResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &lb.UpdateListenerResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}

func (c *LbClient) UpdateListenerCertificates(request *lb.UpdateListenerCertificatesRequest) (*lb.UpdateListenerCertificatesResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &lb.UpdateListenerCertificatesResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}
