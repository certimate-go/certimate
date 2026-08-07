package safeline

type sdkResponse interface {
	GetErrCode() string
	GetErrMsg() string
}

type sdkResponseBase struct {
	ErrCode *string `json:"err,omitempty"`
	ErrMsg  *string `json:"msg,omitempty"`
}

func (r *sdkResponseBase) GetErrCode() string {
	if r.ErrCode == nil {
		return ""
	}

	return *r.ErrCode
}

func (r *sdkResponseBase) GetErrMsg() string {
	if r.ErrMsg == nil {
		return ""
	}

	return *r.ErrMsg
}

var _ sdkResponse = (*sdkResponseBase)(nil)
