package synologydsm

type sdkResponse interface {
	GetSuccess() bool
	GetErrorCode() int
}

type sdkResponseBase struct {
	Success bool `json:"success"`
	Error   *struct {
		Code int `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

func (r *sdkResponseBase) GetSuccess() bool {
	return r.Success
}

func (r *sdkResponseBase) GetErrorCode() int {
	if r.Error == nil {
		if r.Success {
			return 0
		}
		return -1
	}

	return r.Error.Code
}

var _ sdkResponse = (*sdkResponseBase)(nil)
