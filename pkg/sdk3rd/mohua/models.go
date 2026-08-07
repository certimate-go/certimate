package mohua

type DomainInfo struct {
	ID        int    `json:"id"`
	HostID    int    `json:"host_id"`
	UID       int    `json:"uid"`
	Domain    string `json:"domain"`
	SSLCertID int    `json:"ssl_cert_id"`
	SSLForce  int    `json:"ssl_force"`
}
