package faas

type CustomDomain struct {
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
