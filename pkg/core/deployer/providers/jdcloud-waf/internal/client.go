package internal

import (
	"encoding/json"
	"errors"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	waf "github.com/jdcloud-api/jdcloud-sdk-go/services/waf/apis"
)

// This is a partial copy of https://github.com/jdcloud-api/jdcloud-sdk-go/blob/master/services/waf/client/WafClient.go
// to lightweight the vendor packages in the built binary.
type WafClient struct {
	core.JDCloudClient
}

func NewWafClient(credential *core.Credential) *WafClient {
	if credential == nil {
		return nil
	}

	config := core.NewConfig()
	config.SetEndpoint("waf.jdcloud-api.com")

	return &WafClient{
		core.JDCloudClient{
			Credential:  *credential,
			Config:      *config,
			ServiceName: "waf",
			Revision:    "1.0.9",
			Logger:      core.NewDummyLogger(),
		},
	}
}

func (c *WafClient) BindCert(request *waf.BindCertRequest) (*waf.BindCertResponse, error) {
	if request == nil {
		return nil, errors.New("Request object is nil.")
	}
	resp, err := c.Send(request, c.ServiceName)
	if err != nil {
		return nil, err
	}

	jdResp := &waf.BindCertResponse{}
	err = json.Unmarshal(resp, jdResp)
	if err != nil {
		c.Logger.Log(core.LogError, "Unmarshal json failed, resp: %s", string(resp))
		return nil, err
	}

	return jdResp, err
}
