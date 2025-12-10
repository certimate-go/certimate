package baishan

type apiResponse interface {
	GetStatus() int
	GetMessages() []string
	GetWarnings() []string
	GetErrors() []string
}

type apiResponseBase struct {
	Metadata struct {
		Transformed int `json:"transformed,omitempty"`
	} `json:"metadata"`
	Status   int      `json:"status,omitempty"`
	Messages []string `json:"messages,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

func (r *apiResponseBase) GetStatus() int {
	return r.Status
}

func (r *apiResponseBase) GetMessages() []string {
	return r.Messages
}

func (r *apiResponseBase) GetWarnings() []string {
	return r.Warnings
}

func (r *apiResponseBase) GetErrors() []string {
	return r.Errors
}

var _ apiResponse = (*apiResponseBase)(nil)
