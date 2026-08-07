package dynv6

type ZoneRecord struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	IPv4Address string `json:"ipv4address"`
	IPv6Prefix  string `json:"ipv6prefix"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type DNSRecord struct {
	ID           int64  `json:"id"`
	ZoneID       int64  `json:"zoneID"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Port         int    `json:"port"`
	Weight       int    `json:"weight"`
	Priority     int    `json:"priority"`
	Data         string `json:"data"`
	ExpandedData string `json:"expandedData"`
	Flags        int    `json:"flags,omitempty"`
	Tag          string `json:"tag,omitempty"`
}
