package ecloudsdkcore

type HttpRequest struct {
	Url          string                 `json:"url,omitempty"`
	DefaultUrl   string                 `json:"defaultUrl,omitempty"`
	Method       string                 `json:"method,omitempty"`
	Action       string                 `json:"action,omitempty"`
	Product      string                 `json:"product,omitempty"`
	Version      string                 `json:"version,omitempty"`
	SdkVersion   string                 `json:"sdkVersion,omitempty"`
	Body         interface{}            `json:"body,omitempty"`
	PathParams   map[string]interface{} `json:"pathParams,omitempty"`
	QueryParams  map[string]interface{} `json:"queryParams,omitempty"`
	HeaderParams map[string]interface{} `json:"headerParams,omitempty"`
}

func NewDefaultHttpRequest() *HttpRequest {
	return &HttpRequest{
		DefaultUrl: "https://ecloud.10086.cn",
		Method:     "POST",
	}
}
