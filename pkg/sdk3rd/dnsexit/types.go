package dnsexit

type apiResponse interface {
	GetCode() int32
	GetMessage() string
}

type apiResponseBase struct {
	Code    *int32  `json:"code,omitempty"`
	Message *string `json:"message,omitempty"`
}

func (r *apiResponseBase) GetCode() int32 {
	if r.Code == nil {
		return 0
	}

	return *r.Code
}

func (r *apiResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}

	return *r.Message
}

var _ apiResponse = (*apiResponseBase)(nil)

type DnsRecord struct {
	Type      *string `json:"type,omitempty"`
	Name      *string `json:"name,omitempty"`
	Content   *string `json:"content,omitempty"`
	TTL       *int    `json:"ttl,omitempty"`
	Priority  *int    `json:"priority,omitempty"`
	Overwrite *bool   `json:"overwrite,omitempty"`
}
