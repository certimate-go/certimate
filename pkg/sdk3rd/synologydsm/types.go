package synologydsm

type sdkResponse interface {
	GetSuccess() bool
	GetErrorCode() int
}

type sdkResponseBase struct {
	Success bool      `json:"success"`
	Error   *APIError `json:"error,omitempty"`
}

func (r *sdkResponseBase) GetSuccess() bool {
	return r.Success
}

func (r *sdkResponseBase) GetErrorCode() int {
	if r.Error == nil {
		if r.Success {
			return 0
		}
		return -1
	}

	return r.Error.Code
}

var _ sdkResponse = (*sdkResponseBase)(nil)

type APIError struct {
	Code int `json:"code,omitempty"`
}

type APIInfo struct {
	Path       string `json:"path"`
	MinVersion int    `json:"minVersion"`
	MaxVersion int    `json:"maxVersion"`
}

type CertificateInfo struct {
	ID          string `json:"id"`
	Description string `json:"desc"`
	IsDefault   bool   `json:"is_default"`
	IsBroken    bool   `json:"is_broken"`
	Issuer      struct {
		CommonName   string `json:"common_name"`
		Country      string `json:"country"`
		Organization string `json:"organization"`
	} `json:"issuer"`
	Subject struct {
		CommonName   string   `json:"common_name"`
		Country      string   `json:"country"`
		Organization string   `json:"organization"`
		SAN          []string `json:"sub_alt_name"`
	} `json:"subject"`
	ValidFrom          string                `json:"valid_from"`
	ValidTill          string                `json:"valid_till"`
	SignatureAlgorithm string                `json:"signature_algorithm"`
	Renewable          bool                  `json:"renewable"`
	Services           []*CertificateService `json:"services"`
}

type CertificateService struct {
	DisplayName     string `json:"display_name"`
	DisplayNameI18N string `json:"display_name_i18n,omitempty"`
	IsPkg           bool   `json:"isPkg"`
	Owner           string `json:"owner"`
	Service         string `json:"service"`
	Subscriber      string `json:"subscriber"`
}
