package samwaf

type SSLConfig struct {
	Id          string `json:"id"`
	CertContent string `json:"cert_content"`
	KeyContent  string `json:"key_content"`
	SerialNo    string `json:"serial_no"`
	Subject     string `json:"subject"`
	Issuer      string `json:"issuer"`
	ValidFrom   string `json:"valid_from"`
	ValidTo     string `json:"valid_to"`
	Domains     string `json:"domains"`
	KeyPath     string `json:"key_path"`
	CertPath    string `json:"cert_path"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
}
