package domain

const ACMEChallengeDNSPersist01 = "dns-persist-01"

type DNSPersistRecord struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

type DNSPersistRecordStatus struct {
	DNSPersistRecord
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
