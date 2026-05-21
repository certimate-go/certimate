package dynadot

type sdkResponse interface {
	GetCode() int
	GetMessage() string
}

type sdkResponseBase struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r *sdkResponseBase) GetCode() int {
	return r.Code
}

func (r *sdkResponseBase) GetMessage() string {
	return r.Message
}

var _ sdkResponse = (*sdkResponseBase)(nil)

type DnsMainRecord struct {
	RecordType   string `json:"record_type"`
	RecordValue1 string `json:"record_value1"`
	RecordValue2 string `json:"record_value2,omitempty"`
}

type DnsSubRecord struct {
	SubHost      string `json:"sub_host"`
	RecordType   string `json:"record_type"`
	RecordValue1 string `json:"record_value1"`
	RecordValue2 string `json:"record_value2,omitempty"`
}
