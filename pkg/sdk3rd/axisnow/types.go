package axisnow

import (
	"fmt"
	"strings"
)

type sdkResponse interface {
	GetAPIError() error
	GetSuccess() bool
}

type sdkResponseBase struct {
	Success bool         `json:"success"`
	Errors  sdkAPIErrors `json:"errors,omitempty"`
}

type sdkAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type sdkAPIErrors []sdkAPIError

func (e sdkAPIErrors) Error() string {
	builder := &strings.Builder{}

	for _, item := range e {
		fmt.Fprintf(builder, "[%d] %s", item.Code, item.Message)
	}

	return strings.TrimSpace(builder.String())
}

func (r *sdkResponseBase) GetSuccess() bool {
	return r.Success
}

func (r *sdkResponseBase) GetAPIError() error {
	if len(r.Errors) > 0 {
		return r.Errors
	}
	return nil
}

var _ sdkResponse = (*sdkResponseBase)(nil)
