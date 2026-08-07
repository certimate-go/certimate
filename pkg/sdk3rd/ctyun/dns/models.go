package dns

type DNSRecord struct {
	RecordId int32  `json:"recordId"`
	Host     string `json:"host"`
	Type     string `json:"type"`
	LineCode string `json:"lineCode"`
	Value    string `json:"value"`
	TTL      int32  `json:"ttl"`
	State    int32  `json:"state"`
	Remark   string `json:"remark"`
}
