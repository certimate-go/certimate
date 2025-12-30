package faas

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type apiResponse interface {
	GetStatusCode() string
	GetMessage() string
	GetError() string
	GetErrorMessage() string
}

type apiResponseBase struct {
	StatusCode   json.RawMessage `json:"statusCode,omitempty"`
	Message      *string         `json:"message,omitempty"`
	Error        *string         `json:"error,omitempty"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
	RequestId    *string         `json:"requestId,omitempty"`
}

func (r *apiResponseBase) GetStatusCode() string {
	if r.StatusCode == nil {
		return ""
	}

	decoder := json.NewDecoder(bytes.NewReader(r.StatusCode))
	token, err := decoder.Token()
	if err != nil {
		return ""
	}

	switch t := token.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func (r *apiResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}

	return *r.Message
}

func (r *apiResponseBase) GetError() string {
	if r.Error == nil {
		return ""
	}

	return *r.Error
}

func (r *apiResponseBase) GetErrorMessage() string {
	if r.ErrorMessage == nil {
		return ""
	}

	return *r.ErrorMessage
}

var _ apiResponse = (*apiResponseBase)(nil)

type CustomDomainRecord struct {
	DomainName   string                   `json:"domainName"`
	Protocol     string                   `json:"protocol"`
	AuthConfig   *CustomDomainAuthConfig  `json:"authConfig,omitempty"`
	CertConfig   *CustomDomainCertConfig  `json:"certConfig,omitempty"`
	RouteConfig  *CustomDomainRouteConfig `json:"routeConfig,omitempty"`
	DomainStatus string                   `json:"domainStatus"`
	CnameValid   bool                     `json:"cnameValid"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
}

type CustomDomainAuthConfig struct {
	AuthType  string                     `json:"authType"`
	JwtConfig *CustomDomainAuthJwtConfig `json:"jwtConfig,omitempty"`
}

type CustomDomainAuthJwtConfig struct {
	Jwks        string                              `json:"jwks"`
	TokenConfig *CustomDomainAuthJwtTokenConfig     `json:"tokenConfig,omitempty"`
	ClaimTrans  []*CustomDomainAuthJwtClaimTran     `json:"claimTrans,omitempty"`
	MatchMode   *CustomDomainAuthJwtMatchModeConfig `json:"matchMode,omitempty"`
}

type CustomDomainAuthJwtClaimTran struct {
	ClaimName     string `json:"claimName"`
	TargetName    string `json:"targetName"`
	TransLocation string `json:"transLocation"`
}

type CustomDomainAuthJwtTokenConfig struct {
	Location     string  `json:"location"`
	Name         string  `json:"name"`
	RemovePrefix *string `json:"removePrefix,omitempty"`
}

type CustomDomainAuthJwtMatchModeConfig struct {
	Mode string   `json:"mode"`
	Path []string `json:"path"`
}

type CustomDomainCertConfig struct {
	CertName    string `json:"certName"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
}

type CustomDomainRouteConfig struct {
	Routes []*CustomDomainRoutePathConfig `json:"routes"`
}

type CustomDomainRoutePathConfig struct {
	EnableJwt          int32    `json:"enableJwt"`
	FunctionId         int64    `json:"functionId"`
	FunctionName       string   `json:"functionName"`
	FunctionUniqueName string   `json:"functionUniqueName"`
	Methods            []string `json:"methods"`
	Path               string   `json:"path"`
	Qualifier          string   `json:"qualifier,omitempty"`
}
