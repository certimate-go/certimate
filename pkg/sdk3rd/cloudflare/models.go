package cloudflare

type GeoRestriction struct {
	Label string `json:"label"`
}

type CustomCertificate struct {
	ID                 string           `json:"id"`
	ZoneID             string           `json:"zone_id"`
	BundleMethod       string           `json:"bundle_method"`
	CustomCsrID        string           `json:"custom_csr_id"`
	GeoRestrictions    []GeoRestriction `json:"geo_restrictions"`
	Hosts              []string         `json:"hosts"`
	Issuer             string           `json:"issuer"`
	PolicyRestrictions string           `json:"policy_restrictions"`
	Priority           float64          `json:"priority"`
	Signature          string           `json:"signature"`
	Status             string           `json:"status"`
	ExpiresOn          string           `json:"expires_on"`
	UploadedOn         string           `json:"uploaded_on"`
	ModifiedOn         string           `json:"modified_on"`
}
