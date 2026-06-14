package digitalocean

type sdkResponse interface {
	GetId() string
	GetMessage() string
}

type sdkResponseBase struct {
	Id      *string `json:"id,omitempty"`
	Message *string `json:"message,omitempty"`
}

func (r *sdkResponseBase) GetId() string {
	if r.Id == nil {
		return ""
	}

	return *r.Id
}

func (r *sdkResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}

	return *r.Message
}

var _ sdkResponse = (*sdkResponseBase)(nil)
