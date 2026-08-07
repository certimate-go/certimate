package dokploy

type Certificate struct {
	CertificateId   string `json:"certificateId"`
	Name            string `json:"name"`
	CertificateData string `json:"certificateData"`
	PrivateKey      string `json:"privateKey"`
	CertificatePath string `json:"certificatePath,omitempty"`
	OrganizationId  string `json:"organizationId,omitempty"`
	ServerId        string `json:"serverId,omitempty"`
}
