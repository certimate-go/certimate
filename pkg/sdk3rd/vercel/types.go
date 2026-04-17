package vercel

type sdkResponse interface {
	GetError() *sdkError
}

type sdkResponseBase struct {
	Error *sdkError `json:"error,omitempty"`
}

type sdkError struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	MissingToken bool   `json:"missingToken,omitempty"`
}

func (r *sdkResponseBase) GetError() *sdkError {
	return r.Error
}

var _ sdkResponse = (*sdkResponseBase)(nil)
