package ecloudsdkcore

type OpenApiRequest struct {
	Product         string                 `json:"product,omitempty"`
	Version         string                 `json:"version,omitempty"`
	SdkVersion      string                 `json:"sdkVersion,omitempty"`
	Language        string                 `json:"language,omitempty"`
	Api             string                 `json:"api,omitempty"`
	PoolId          string                 `json:"poolId,omitempty"`
	HeaderParameter map[string]interface{} `json:"headerParameter,omitempty"`
	PathParameter   map[string]interface{} `json:"pathParameter,omitempty"`
	QueryParameter  map[string]interface{} `json:"queryParameter,omitempty"`
	BodyParameter   interface{}            `json:"bodyParameter,omitempty"`
	AccessKey       string                 `json:"accessKey,omitempty"`
	SecretKey       string                 `json:"secretKey,omitempty"`
}
