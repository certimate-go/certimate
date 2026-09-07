package v4

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
