package synology

// APIInfo represents API information from the query response.
type APIInfo struct {
	Path       string `json:"path"`
	MinVersion int    `json:"minVersion"`
	MaxVersion int    `json:"maxVersion"`
}

// QueryAPIInfoResponse represents the response from the API info query.
type QueryAPIInfoResponse struct {
	Success bool               `json:"success"`
	Data    map[string]APIInfo `json:"data,omitempty"`
	Error   *APIError          `json:"error,omitempty"`
}

// APIError represents an error from the Synology API.
type APIError struct {
	Code int `json:"code"`
}

// LoginData represents the data returned from a successful login.
type LoginData struct {
	Sid       string `json:"sid"`
	SynoToken string `json:"synotoken"`
	DeviceId  string `json:"device_id,omitempty"`
	Did       string `json:"did,omitempty"`
}

// LoginResponse represents the response from the login API.
type LoginResponse struct {
	Success bool       `json:"success"`
	Data    *LoginData `json:"data,omitempty"`
	Error   *APIError  `json:"error,omitempty"`
}

// CertificateService represents a service associated with a certificate.
type CertificateService struct {
	DisplayName     string `json:"display_name"`
	DisplayNameI18N string `json:"display_name_i18n,omitempty"`
	IsPkg           bool   `json:"isPkg"`
	Owner           string `json:"owner"`
	Service         string `json:"service"`
	Subscriber      string `json:"subscriber"`
	MultipleCert    bool   `json:"multiple_cert,omitempty"`
	UserSettable    bool   `json:"user_setable,omitempty"`
}

// Certificate represents a certificate in Synology DSM.
type Certificate struct {
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
		SubAltName   []string `json:"sub_alt_name"`
		Country      string   `json:"country"`
		Organization string   `json:"organization"`
	} `json:"subject"`
	ValidFrom          string               `json:"valid_from"`
	ValidTill          string               `json:"valid_till"`
	Services           []CertificateService `json:"services"`
	Renewable          bool                 `json:"renewable"`
	SignatureAlgorithm string               `json:"signature_algorithm"`
}

// ListCertificatesData represents the data returned from the list certificates API.
type ListCertificatesData struct {
	Certificates []Certificate `json:"certificates"`
}

// ListCertificatesResponse represents the response from the list certificates API.
type ListCertificatesResponse struct {
	Success bool                  `json:"success"`
	Data    *ListCertificatesData `json:"data,omitempty"`
	Error   *APIError             `json:"error,omitempty"`
}

// ImportCertificateData represents the data returned from the import certificate API.
type ImportCertificateData struct {
	RestartHttpd bool `json:"restart_httpd"`
}

// ImportCertificateResponse represents the response from the import certificate API.
type ImportCertificateResponse struct {
	Success bool                   `json:"success"`
	Data    *ImportCertificateData `json:"data,omitempty"`
	Error   *APIError              `json:"error,omitempty"`
}

// CertificateWithServices represents a certificate with its associated services.
type CertificateWithServices struct {
	ID          string               `json:"id"`
	Description string               `json:"desc"`
	IsDefault   bool                 `json:"is_default"`
	Services    []CertificateService `json:"services"`
}

// ListCertificatesWithServicesData represents the data returned from the list certificates API with services.
type ListCertificatesWithServicesData struct {
	Certificates []CertificateWithServices `json:"certificates"`
}

// ListCertificatesWithServicesResponse represents the response from the list certificates API with services.
type ListCertificatesWithServicesResponse struct {
	Success bool                              `json:"success"`
	Data    *ListCertificatesWithServicesData `json:"data,omitempty"`
	Error   *APIError                         `json:"error,omitempty"`
}

// ServiceCertificateSetting represents a setting to assign a certificate to a service.
type ServiceCertificateSetting struct {
	Service CertificateService `json:"service"`
	OldID   string             `json:"old_id"`
	ID      string             `json:"id"`
}

// SetServiceCertificateResponse represents the response from setting service certificate.
type SetServiceCertificateResponse struct {
	Success bool      `json:"success"`
	Error   *APIError `json:"error,omitempty"`
}
