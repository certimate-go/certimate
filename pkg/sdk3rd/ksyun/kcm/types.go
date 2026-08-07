package kcm

import (
	"fmt"
)

type sdkResponse interface {
	GetAPIError() error
}

type sdkResponseBase struct {
	RequestId *string      `json:"RequestId,omitempty"`
	Error     *sdkAPIError `json:"Error,omitempty"`
}

type sdkAPIError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

func (e sdkAPIError) Error() string {
	return fmt.Sprintf("[%s] %s ", e.Code, e.Message)
}

func (r *sdkResponseBase) GetAPIError() error {
	if r.Error != nil {
		return r.Error
	}
	return nil
}

var _ sdkResponse = (*sdkResponseBase)(nil)
