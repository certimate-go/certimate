package cloudflare

import (
	"fmt"
	"strings"
)

type sdkResponse interface {
	GetAPIError() error
	GetSuccess() bool
}

type sdkResponseBase struct {
	Errors   sdkAPIErrors    `json:"errors,omitempty"`
	Messages []sdkAPIMessage `json:"messages,omitempty"`
	Success  bool            `json:"success,omitempty"`
}

type sdkAPIMessage struct {
	Code             int                `json:"code"`
	Message          string             `json:"message"`
	DocumentationURL string             `json:"documentation_url"`
	ErrorChain       []sdkAPIErrorChain `json:"error_chain"`
	Source           *sdkAPISource      `json:"source"`
}

type sdkAPIErrors []sdkAPIMessage

type sdkAPIErrorChain struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type sdkAPISource struct {
	Pointer string `json:"pointer"`
}

func (e sdkAPIErrors) Error() string {
	builder := &strings.Builder{}

	for _, item := range e {
		fmt.Fprintf(builder, "[%d] %s", item.Code, item.Message)

		for _, link := range item.ErrorChain {
			fmt.Fprintf(builder, "; [%d] %s", link.Code, link.Message)
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

func (r *sdkResponseBase) GetSuccess() bool {
	return r.Success
}

var _ sdkResponse = (*sdkResponseBase)(nil)
