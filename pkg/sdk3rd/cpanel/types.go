package cpanel

type sdkResponse interface {
	GetStatus() int
	GetMessages() []string
	GetWarnings() []string
	GetErrors() []string
}

type sdkResponseBase struct {
	Metadata struct {
		Transformed int `json:"transformed,omitempty"`
	} `json:"metadata"`
	Status   int      `json:"status,omitempty"`
	Messages []string `json:"messages,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

func (r *sdkResponseBase) GetStatus() int {
	return r.Status
}

func (r *sdkResponseBase) GetMessages() []string {
	return r.Messages
}

func (r *sdkResponseBase) GetWarnings() []string {
	return r.Warnings
}

func (r *sdkResponseBase) GetErrors() []string {
	return r.Errors
}

var _ sdkResponse = (*sdkResponseBase)(nil)
