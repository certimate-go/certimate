package udnr

type DomainDNSRecord struct {
	Type     string `json:"DnsType,omitempty"`
	Name     string `json:"RecordName,omitempty"`
	Content  string `json:"Content,omitempty"`
	Priority string `json:"Prio,omitempty"`
	TTL      string `json:"TTL,omitempty"`
}
