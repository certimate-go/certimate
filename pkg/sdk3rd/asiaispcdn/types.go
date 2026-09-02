package asiaispcdn

import "encoding/json"

type sdkResponse interface {
	GetCode() int
	GetMsg() string
}

type sdkResponseBase struct {
	Code *json.Number `json:"code,omitempty"`
	Msg  *string      `json:"msg,omitempty"`
}

func (r *sdkResponseBase) GetCode() int {
	if r.Code.String() == "" {
		return 0
	}

	code, err := r.Code.Int64()
	if err != nil {
		return -1
	}

	return int(code)
}

func (r *sdkResponseBase) GetMsg() string {
	if r.Msg == nil {
		return ""
	}

	return *r.Msg
}

var _ sdkResponse = (*sdkResponseBase)(nil)
