package proxmoxve

type CustomCertificate struct {
	Certificates string `json:"certificates,omitempty"`
	Force        bool   `json:"force,omitempty"`
	Key          string `json:"key,omitempty"`
	Restart      bool   `json:"restart,omitempty"`
}
