package dnscom

type apiResponse interface {
	GetCode() int32
	GetMessage() string
}

type apiResponseBase struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (r *apiResponseBase) GetCode() int32 {
	return r.Code
}

func (r *apiResponseBase) GetMessage() string {
	return r.Message
}

var _ apiResponse = (*apiResponseBase)(nil)

type DomainRecord struct {
	DomainID int64  `json:"domainID"`
	Domain   string `json:"domain"`
	State    int32  `json:"state"`
}

type DNSRecord struct {
	DomainID int64  `json:"domainID"`
	RecordID int64  `json:"recordID"`
	ViewID   int64  `json:"viewID"`
	Record   string `json:"record"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Value    string `json:"value"`
	TTL      int32  `json:"ttl"`
	MX       int32  `json:"mx"`
	State    int32  `json:"state"`
	Remark   string `json:"remark"`
}
