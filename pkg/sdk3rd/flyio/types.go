package flyio

type sdkResponse interface {
	GetError() string
}

type sdkResponseBase struct {
	Error *string `json:"error,omitempty"`
}

func (r *sdkResponseBase) GetError() string {
	if r.Error == nil {
		return ""
	}

	return *r.Error
}

var _ sdkResponse = (*sdkResponseBase)(nil)
