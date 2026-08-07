package linode

import (
	"fmt"
	"strings"
)

type sdkResponse interface {
	GetAPIError() error
}

type sdkResponseBase struct {
	Errors sdkAPIErrors `json:"errors,omitempty"`
}

type sdkAPIErrorReason struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason"`
}

type sdkAPIErrors []sdkAPIErrorReason

func (e sdkAPIErrors) Error() string {
	builder := &strings.Builder{}

	for _, item := range e {
		if item.Field == "" {
			fmt.Fprintf(builder, "%s ", item.Reason)
		} else {
			fmt.Fprintf(builder, "[%s] %s ", item.Field, item.Reason)
		}
	}

	return strings.TrimSpace(builder.String())
}

func (r *sdkResponseBase) GetAPIError() error {
	if len(r.Errors) > 0 {
		return r.Errors
	}
	return nil
}

var _ sdkResponse = (*sdkResponseBase)(nil)
