package v3

type sdkResponse interface {
	GetCode() int
	GetError() string
}

type sdkResponseBase struct {
	Code  *int    `json:"code,omitempty"`
	Error *string `json:"error,omitempty"`
}

func (r *sdkResponseBase) GetCode() int {
	if r.Code == nil {
		return 0
	}

	return *r.Code
}

func (r *sdkResponseBase) GetError() string {
	if r.Error == nil {
		return ""
	}

	return *r.Error
}

var _ sdkResponse = (*sdkResponseBase)(nil)
