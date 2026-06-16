package mohua

type sdkResponse interface {
	GetStatus() int
	GetMsg() string
}

type sdkResponseBase struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

func (r *sdkResponseBase) GetStatus() int {
	return r.Status
}

func (r *sdkResponseBase) GetMsg() string {
	return r.Msg
}

var _ sdkResponse = (*sdkResponseBase)(nil)
