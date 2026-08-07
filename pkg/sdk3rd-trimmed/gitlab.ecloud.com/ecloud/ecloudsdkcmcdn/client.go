package ecloudsdkcmcdn

import (
	"gitlab.ecloud.com/ecloud/ecloudsdkcmcdn/model"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/param"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/request"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/utils"
)

type Client struct {
	apiClient   *ecloudsdkcore.APIClient
	config      *config.Config
	httpRequest *request.HttpRequest
	allRegions  map[string]string
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
	product    string = "CM-CDN"
	version           = "v1"
	sdkVersion        = "1.0.0"
)

func (c *Client) initRegions() map[string]string {
	m := map[string]string{
		"CIDC-RP-04":   "https://console-yunnan-1.cmecloud.cn:8443",
		"CIDC-RP-16":   "https://console-qinghai-1.cmecloud.cn:8443",
		"CIDC-RP-25":   "https://console-wuxi-1.cmecloud.cn:8443",
		"CIDC-RP-26":   "https://console-dongguan-1.cmecloud.cn:8443",
		"CIDC-RP-27":   "https://console-yaan-1.cmecloud.cn:8443",
		"CIDC-RP-28":   "https://console-zhengzhou-1.cmecloud.cn:8443",
		"CIDC-RP-29":   "https://console-beijing-2.cmecloud.cn:8443",
		"CIDC-RP-30":   "https://console-zhuzhou-1.cmecloud.cn:8443",
		"CIDC-RP-31":   "https://console-jinan-1.cmecloud.cn:8443",
		"CIDC-RP-32":   "https://console-xian-1.cmecloud.cn:8443",
		"CIDC-RP-33":   "https://console-shanghai-1.cmecloud.cn:8443",
		"CIDC-RP-34":   "https://console-chongqing-1.cmecloud.cn:8443",
		"CIDC-RP-35":   "https://console-ningbo-1.cmecloud.cn:8443",
		"CIDC-RP-36":   "https://console-tianjin-1.cmecloud.cn:8443",
		"CIDC-RP-37":   "https://console-jilin-1.cmecloud.cn:8443",
		"CIDC-RP-38":   "https://console-hubei-1.cmecloud.cn:8443",
		"CIDC-RP-39":   "https://console-jiangxi-1.cmecloud.cn:8443",
		"CIDC-RP-40":   "https://console-gansu-1.cmecloud.cn:8443",
		"CIDC-RP-41":   "https://console-shanxi-1.cmecloud.cn:8443",
		"CIDC-RP-42":   "https://console-liaoning-1.cmecloud.cn:8443",
		"CIDC-RP-43":   "https://console-yunnan-2.cmecloud.cn:8443",
		"CIDC-RP-44":   "https://console-hebei-1.cmecloud.cn:8443",
		"CIDC-RP-45":   "https://console-fujian-1.cmecloud.cn:8443",
		"CIDC-RP-46":   "https://console-guangxi-1.cmecloud.cn:8443",
		"CIDC-RP-47":   "https://console-anhui-1.cmecloud.cn:8443",
		"CIDC-RP-48":   "https://console-huhehaote-1.cmecloud.cn:8443",
		"CIDC-RP-49":   "https://console-guiyang-1.cmecloud.cn:8443",
		"CIDC-CORE-00": "https://ecloud.10086.cn",
		"CIDC-RP-53":   "https://console-hainan-1.cmecloud.cn:8443",
		"CIDC-RP-54":   "https://console-xinjiang-1.cmecloud.cn:8443",
		"CIDC-RP-55":   "http://console-heilongjiang-1.cmecloud.cn:18080",
		"CIDC-BRP-25":  "",
		"CIDC-RP-60":   "",
		"CIDC-RP-61":   "",
		"CIDC-RP-62":   "",
	}
	return m
}

func (c *Client) setEndpoint(config *config.Config, httpRequest *request.HttpRequest) {
	if utils.IsUnSet(config.PoolId) {
		httpRequest.Endpoint = utils.DefaultEndpoint
		return
	}
	endpoint := c.allRegions[*config.PoolId]
	if utils.IsUnSet(endpoint) {
		httpRequest.Endpoint = utils.DefaultEndpoint
		return
	}
	httpRequest.Endpoint = endpoint
}

func (c *Client) DescribeCdnCertificateDetail(request *model.DescribeCdnCertificateDetailRequest) (*model.DescribeCdnCertificateDetailResponse, error) {
	return c.DescribeCdnCertificateDetailWithConfig(request, nil)
}

func (c *Client) DescribeCdnCertificateDetailWithConfig(request *model.DescribeCdnCertificateDetailRequest, runtimeConfig *config.RuntimeConfig) (*model.DescribeCdnCertificateDetailResponse, error) {
	params := param.NewParamsBuilder().
		Action("describeCdnCertificateDetail").
		Uri("/domainManager/openapi/certificate/describeCdnCertificateDetail/{uniqueId}").
		GatewayUri("/api/openapi-ecdn/domainManager/openapi/certificate/describeCdnCertificateDetail/{uniqueId}").
		Protocol("https").
		Method("GET").
		Request(request).
		Build()
	returnValue := &model.DescribeCdnCertificateDetailResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) DescribeCdnDomainDetail(request *model.DescribeCdnDomainDetailRequest) (*model.DescribeCdnDomainDetailResponse, error) {
	return c.DescribeCdnDomainDetailWithConfig(request, nil)
}

func (c *Client) DescribeCdnDomainDetailWithConfig(request *model.DescribeCdnDomainDetailRequest, runtimeConfig *config.RuntimeConfig) (*model.DescribeCdnDomainDetailResponse, error) {
	params := param.NewParamsBuilder().
		Action("describeCdnDomainDetail").
		Uri("/domainManager/openapi/domain/describeCdnDomainDetail/{domainId}").
		GatewayUri("/api/openapi-ecdn/domainManager/openapi/domain/describeCdnDomainDetail/{domainId}").
		Protocol("https").
		Method("GET").
		Request(request).
		Build()
	returnValue := &model.DescribeCdnDomainDetailResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) DescribeUserDomains(request *model.DescribeUserDomainsRequest) (*model.DescribeUserDomainsResponse, error) {
	return c.DescribeUserDomainsWithConfig(request, nil)
}

func (c *Client) DescribeUserDomainsWithConfig(request *model.DescribeUserDomainsRequest, runtimeConfig *config.RuntimeConfig) (*model.DescribeUserDomainsResponse, error) {
	params := param.NewParamsBuilder().
		Action("describeUserDomains").
		Uri("/domainManager/openapi/domain/describeUserDomains").
		GatewayUri("/api/openapi-ecdn/domainManager/openapi/domain/describeUserDomains").
		Protocol("https").
		Method("GET").
		Request(request).
		Build()
	returnValue := &model.DescribeUserDomainsResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}

func (c *Client) AddDomainServerCertificate(request *model.AddDomainServerCertificateRequest) (*model.AddDomainServerCertificateResponse, error) {
	return c.AddDomainServerCertificateWithConfig(request, nil)
}

func (c *Client) AddDomainServerCertificateWithConfig(request *model.AddDomainServerCertificateRequest, runtimeConfig *config.RuntimeConfig) (*model.AddDomainServerCertificateResponse, error) {
	params := param.NewParamsBuilder().
		Action("addDomainServerCertificate").
		Uri("/domainManager/openapi/certificate/addDomainServerCertificate").
		GatewayUri("/api/openapi-ecdn/domainManager/openapi/certificate/addDomainServerCertificate").
		Protocol("https").
		ContentType("application/json").
		Method("POST").
		Request(request).
		Build()
	returnValue := &model.AddDomainServerCertificateResponse{}
	if _, err := c.apiClient.Excute(params, runtimeConfig, returnValue); err != nil {
		return nil, err
	} else {
		return returnValue, nil
	}
}
