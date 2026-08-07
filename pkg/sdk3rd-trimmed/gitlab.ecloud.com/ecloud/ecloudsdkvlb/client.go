package ecloudsdkvlb

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcore"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/param"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/region"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/request"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/utils"
	"gitlab.ecloud.com/ecloud/ecloudsdkvlb/model"
)

type Client struct {
	apiClient   *ecloudsdkcore.APIClient
	config      *config.Config
	httpRequest *request.HttpRequest
	allRegions  []region.Region
}

func NewClient(config *config.Config) *Client {
	httpRequest := request.DefaultHttpRequest()
	httpRequest.Product = product
	httpRequest.Version = version
	httpRequest.SdkVersion = sdkVersion
	ecloudsdkcore.InitConfig(config)
	apiClient := ecloudsdkcore.DefaultApiClient(config, httpRequest)
	client := &Client{
		apiClient:   apiClient,
		config:      config,
		httpRequest: httpRequest,
	}
	client.allRegions = client.initRegions()
	client.setEndpoint(config, httpRequest)
	return client
}

const (
	product    string = "vlb"
	version           = "v1"
	sdkVersion        = "1.0.7"
)

func (c *Client) initRegions() []region.Region {
	var regions []region.Region
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-jiangsu-1"),
		PoolId:   utils.String("CIDC-RP-25"),
		Endpoint: utils.String("https://api-wuxi-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-guangdong-1"),
		PoolId:   utils.String("CIDC-RP-26"),
		Endpoint: utils.String("https://api-dongguan-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-sichuan-1"),
		PoolId:   utils.String("CIDC-RP-27"),
		Endpoint: utils.String("https://api-yaan-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-henan-1"),
		PoolId:   utils.String("CIDC-RP-28"),
		Endpoint: utils.String("https://api-zhengzhou-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-beijing-1"),
		PoolId:   utils.String("CIDC-RP-29"),
		Endpoint: utils.String("https://api-beijing-2.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-hunan-1"),
		PoolId:   utils.String("CIDC-RP-30"),
		Endpoint: utils.String("https://api-zhuzhou-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-shandong-1"),
		PoolId:   utils.String("CIDC-RP-31"),
		Endpoint: utils.String("https://api-jinan-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-shaanxi-1"),
		PoolId:   utils.String("CIDC-RP-32"),
		Endpoint: utils.String("https://api-xian-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-shanghai-1"),
		PoolId:   utils.String("CIDC-RP-33"),
		Endpoint: utils.String("https://api-shanghai-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-chongqing-1"),
		PoolId:   utils.String("CIDC-RP-34"),
		Endpoint: utils.String("https://api-chongqing-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-zhejiang-1"),
		PoolId:   utils.String("CIDC-RP-35"),
		Endpoint: utils.String("https://api-ningbo-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-tianjin-1"),
		PoolId:   utils.String("CIDC-RP-36"),
		Endpoint: utils.String("https://api-tianjin-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-jilin-1"),
		PoolId:   utils.String("CIDC-RP-37"),
		Endpoint: utils.String("https://api-jilin-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-hubei-1"),
		PoolId:   utils.String("CIDC-RP-38"),
		Endpoint: utils.String("https://api-hubei-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-jiangxi-1"),
		PoolId:   utils.String("CIDC-RP-39"),
		Endpoint: utils.String("https://api-jiangxi-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-gansu-1"),
		PoolId:   utils.String("CIDC-RP-40"),
		Endpoint: utils.String("https://api-gansu-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-shangxi-1"),
		PoolId:   utils.String("CIDC-RP-41"),
		Endpoint: utils.String("https://api-shanxi-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-liaoning-1"),
		PoolId:   utils.String("CIDC-RP-42"),
		Endpoint: utils.String("https://api-liaoning-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-yunnan-1"),
		PoolId:   utils.String("CIDC-RP-43"),
		Endpoint: utils.String("https://api-yunnan-2.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-hebei-1"),
		PoolId:   utils.String("CIDC-RP-44"),
		Endpoint: utils.String("https://api-hebei-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-fujian-1"),
		PoolId:   utils.String("CIDC-RP-45"),
		Endpoint: utils.String("https://api-fujian-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-guangxi-1"),
		PoolId:   utils.String("CIDC-RP-46"),
		Endpoint: utils.String("https://api-guangxi-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-anhui-1"),
		PoolId:   utils.String("CIDC-RP-47"),
		Endpoint: utils.String("https://api-anhui-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-neimenggu-1"),
		PoolId:   utils.String("CIDC-RP-48"),
		Endpoint: utils.String("https://api-huhehaote-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-guzhou-1"),
		PoolId:   utils.String("CIDC-RP-49"),
		Endpoint: utils.String("https://api-guiyang-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-hainan-1"),
		PoolId:   utils.String("CIDC-RP-53"),
		Endpoint: utils.String("https://api-hainan-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-xinjiang-1"),
		PoolId:   utils.String("CIDC-RP-54"),
		Endpoint: utils.String("https://api-xinjiang-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-heilongjiang-1"),
		PoolId:   utils.String("CIDC-RP-55"),
		Endpoint: utils.String("https://api-heilongjiang-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-ningxia-1"),
		PoolId:   utils.String("CIDC-RP-60"),
		Endpoint: utils.String("https://api-ningxia-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-qinghai-1"),
		PoolId:   utils.String("CIDC-RP-61"),
		Endpoint: utils.String("https://api-qinghai-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-xizang-1"),
		PoolId:   utils.String("CIDC-RP-62"),
		Endpoint: utils.String("https://api-xizang-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-hubei-2"),
		PoolId:   utils.String("CIDC-RP-64"),
		Endpoint: utils.String("https://api-wuhan-1.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-xizang-2"),
		PoolId:   utils.String("CIDC-RP-68"),
		Endpoint: utils.String("https://api-xizang-2.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-qinghai-2"),
		PoolId:   utils.String("CIDC-RP-69"),
		Endpoint: utils.String("https://api-qinghai-2.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String("cn-ningxia-2"),
		PoolId:   utils.String("CIDC-RP-71"),
		Endpoint: utils.String("https://api-ningxia-2.cmecloud.cn:8443"),
	})
	regions = append(regions, region.Region{
		RegionId: utils.String(""),
		PoolId:   utils.String("CIDC-CORE-00"),
		Endpoint: utils.String("https://ecloud.10086.cn"),
	})
	return regions
}

func (c *Client) setEndpoint(config *config.Config, httpRequest *request.HttpRequest) {
	if utils.IsUnSet(config.RegionId) && utils.IsUnSet(config.PoolId) {
		httpRequest.Endpoint = utils.DefaultEndpoint
	} else if utils.IsSet(config.RegionId) {
		c.findAndSetEndpointByRegionId(config.RegionId, httpRequest)
	} else {
		c.findAndSetEndpointByPoolId(config.PoolId, httpRequest)
	}
}

func (c *Client) findAndSetEndpointByRegionId(regionId *string, httpRequest *request.HttpRequest) {
	for _, r := range c.allRegions {
		if utils.StringValue(r.RegionId) == utils.StringValue(regionId) {
			endpoint := r.Endpoint
			if utils.IsUnSet(endpoint) {
				httpRequest.Endpoint = utils.DefaultEndpoint
			} else {
				httpRequest.Endpoint = *endpoint
			}
			break
		}
	}
}

func (c *Client) findAndSetEndpointByPoolId(poolId *string, httpRequest *request.HttpRequest) {
	for _, r := range c.allRegions {
		if utils.StringValue(r.PoolId) == utils.StringValue(poolId) {
			endpoint := r.Endpoint
			if utils.IsUnSet(endpoint) {
				httpRequest.Endpoint = utils.DefaultEndpoint
			} else {
				httpRequest.Endpoint = *endpoint
			}
			break
		}
	}
}

func (c *Client) CreateLoadbalanceCertification(request *model.CreateLoadbalanceCertificationRequest) (*model.CreateLoadbalanceCertificationResponse, error) {
	return c.CreateLoadbalanceCertificationWithConfig(request, nil)
}

func (c *Client) CreateLoadbalanceCertificationWithConfig(request *model.CreateLoadbalanceCertificationRequest, runtimeConfig *config.RuntimeConfig) (*model.CreateLoadbalanceCertificationResponse, error) {
	params := param.NewParamsBuilder().
		Action("createLoadbalanceCertification").
		Uri("/acl/v3/certification").
		GatewayUri("/api/openapi-vlb/lb-console/acl/v3/certification").
		Protocol("http").
		ContentType("application/json").
		Method("POST").
		Request(request).
		Build()
	returnValue := &model.CreateLoadbalanceCertificationResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) GetLoadbalanceCertificationDetailResp(request *model.GetLoadbalanceCertificationDetailRespRequest) (*model.GetLoadbalanceCertificationDetailRespResponse, error) {
	return c.GetLoadbalanceCertificationDetailRespWithConfig(request, nil)
}

func (c *Client) GetLoadbalanceCertificationDetailRespWithConfig(request *model.GetLoadbalanceCertificationDetailRespRequest, runtimeConfig *config.RuntimeConfig) (*model.GetLoadbalanceCertificationDetailRespResponse, error) {
	params := param.NewParamsBuilder().
		Action("getLoadbalanceCertificationDetailResp").
		Uri("/acl/v3/certification/{containerUuid}").
		GatewayUri("/api/openapi-vlb/lb-console/acl/v3/certification/{containerUuid}").
		Protocol("http").
		ContentType("application/json").
		Method("GET").
		Request(request).
		Build()
	returnValue := &model.GetLoadbalanceCertificationDetailRespResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) ListLoadbalanceCertificationResp() (*model.ListLoadbalanceCertificationRespResponse, error) {
	return c.ListLoadbalanceCertificationRespWithConfig(nil)
}

func (c *Client) ListLoadbalanceCertificationRespWithConfig(runtimeConfig *config.RuntimeConfig) (*model.ListLoadbalanceCertificationRespResponse, error) {
	params := param.NewParamsBuilder().
		Action("listLoadbalanceCertificationResp").
		Uri("/acl/v3/certification").
		GatewayUri("/api/openapi-vlb/lb-console/acl/v3/certification").
		Protocol("http").
		ContentType("application/json").
		Method("GET").
		Build()
	returnValue := &model.ListLoadbalanceCertificationRespResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) ListLoadBalanceHTTPSListener(request *model.ListLoadBalanceHTTPSListenerRequest) (*model.ListLoadBalanceHTTPSListenerResponse, error) {
	return c.ListLoadBalanceHTTPSListenerWithConfig(request, nil)
}

func (c *Client) ListLoadBalanceHTTPSListenerWithConfig(request *model.ListLoadBalanceHTTPSListenerRequest, runtimeConfig *config.RuntimeConfig) (*model.ListLoadBalanceHTTPSListenerResponse, error) {
	params := param.NewParamsBuilder().
		Action("listLoadBalanceHTTPSListener").
		Uri("/protocol/v3/listener/{loadBalanceId}/listeners/https").
		GatewayUri("/api/openapi-vlb/lb-console/protocol/v3/listener/{loadBalanceId}/listeners/https").
		Protocol("http").
		ContentType("application/json").
		Method("GET").
		Request(request).
		Build()
	returnValue := &model.ListLoadBalanceHTTPSListenerResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) UpdateListener(request *model.UpdateListenerRequest) (*model.UpdateListenerResponse, error) {
	return c.UpdateListenerWithConfig(request, nil)
}

func (c *Client) UpdateListenerWithConfig(request *model.UpdateListenerRequest, runtimeConfig *config.RuntimeConfig) (*model.UpdateListenerResponse, error) {
	params := param.NewParamsBuilder().
		Action("updateListener").
		Uri("/acl/v3/listener").
		GatewayUri("/api/openapi-vlb/lb-console/acl/v3/listener").
		Protocol("http").
		ContentType("application/json").
		Method("PUT").
		Request(request).
		Build()
	returnValue := &model.UpdateListenerResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}
