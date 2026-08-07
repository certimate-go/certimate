package dns

type DNSRecord struct {
	GroupId     *int64            `json:"group_id,omitempty"`
	GroupStatus *int32            `json:"group_status,omitempty"`
	Value       []*DNSRecordValue `json:"value,omitempty"`
	Weight      *int32            `json:"weight,omitempty"`
}

type DNSRecordValue struct {
	Id    *int64  `json:"id,omitempty"`
	Type  *string `json:"type,omitempty"`
	Value *string `json:"value,omitempty"`
	Line  *string `json:"line,omitempty"`
	Ttl   *int32  `json:"ttl,omitempty"`
}
