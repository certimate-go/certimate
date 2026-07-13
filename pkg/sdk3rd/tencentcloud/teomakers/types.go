package teomakers

type sdkRequest interface {
	SetAction(action string)
}

type sdkRequestBase struct {
	Action *string `json:"Action,omitempty"`
}

func (r *sdkRequestBase) SetAction(action string) {
	r.Action = &action
}

var _ sdkRequest = (*sdkRequestBase)(nil)

type sdkResponse interface {
	GetCode() int
	GetMessage() string
}

type sdkResponseBase struct {
	Code      *int    `json:"Code,omitempty"`
	Message   *string `json:"Message,omitempty"`
	RequestId *string `json:"RequestId,omitempty"`
}

func (r *sdkResponseBase) GetCode() int {
	if r.Code == nil {
		return 0
	}

	return *r.Code
}

func (r *sdkResponseBase) GetMessage() string {
	if r.Message == nil {
		return ""
	}

	return *r.Message
}

var _ sdkResponse = (*sdkResponseBase)(nil)
