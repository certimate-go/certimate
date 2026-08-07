package samwaf

type sdkResponse interface {
	GetCode() int
	GetMsg() string
}

type sdkResponseBase struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func (r *sdkResponseBase) GetCode() int {
	return r.Code
}

func (r *sdkResponseBase) GetMsg() string {
	return r.Msg
}

var _ sdkResponse = (*sdkResponseBase)(nil)
