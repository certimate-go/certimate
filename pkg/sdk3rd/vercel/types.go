package vercel

type sdkResponse interface {
	GetError() *sdkAPIError
}

type sdkResponseBase struct {
	Error *sdkAPIError `json:"error,omitempty"`
}

type sdkAPIError struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	MissingToken bool   `json:"missingToken,omitempty"`
}

func (r *sdkResponseBase) GetError() *sdkAPIError {
	return r.Error
}

var _ sdkResponse = (*sdkResponseBase)(nil)
