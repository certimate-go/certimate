package btwaf

type sdkResponse interface {
	GetCode() int
}

type sdkResponseBase struct {
	Code *int `json:"code,omitempty"`
}

func (r *sdkResponseBase) GetCode() int {
	if r.Code == nil {
		return 0
	}

	return *r.Code
}

var _ sdkResponse = (*sdkResponseBase)(nil)
