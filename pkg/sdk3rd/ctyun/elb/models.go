package elb

type Certificate struct {
	ID          string `json:"ID"`
	RegionID    string `json:"regionID"`
	AzName      string `json:"azName"`
	ProjectID   string `json:"projectID"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
	Status      string `json:"status"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
}

type Listener struct {
	ID                  string `json:"ID"`
	RegionID            string `json:"regionID"`
	AzName              string `json:"azName"`
	ProjectID           string `json:"projectID"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	LoadBalancerID      string `json:"loadBalancerID"`
	Protocol            string `json:"protocol"`
	ProtocolPort        int32  `json:"protocolPort"`
	CertificateID       string `json:"certificateID,omitempty"`
	CaEnabled           bool   `json:"caEnabled"`
	ClientCertificateID string `json:"clientCertificateID,omitempty"`
	Status              string `json:"status"`
	CreatedTime         string `json:"createdTime"`
	UpdatedTime         string `json:"updatedTime"`
}
