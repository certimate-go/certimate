package digitalocean

type Certificate struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name,omitempty"`
	DNSNames        []string `json:"dns_names,omitempty"`
	NotAfter        string   `json:"not_after,omitempty"`
	SHA1Fingerprint string   `json:"sha1_fingerprint,omitempty"`
	Created         string   `json:"created_at,omitempty"`
	State           string   `json:"state,omitempty"`
	Type            string   `json:"type,omitempty"`
}
