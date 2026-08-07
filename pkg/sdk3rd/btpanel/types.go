package btpanel

type sdkResponse interface {
	GetStatus() *bool
	GetMessage() *string
}

type sdkResponseBase struct {
	Status  *bool   `json:"status,omitempty"`
	Message *string `json:"msg,omitempty"`
}

func (r *sdkResponseBase) GetStatus() *bool {
	return r.Status
}

func (r *sdkResponseBase) GetMessage() *string {
	return r.Message
}

var _ sdkResponse = (*sdkResponseBase)(nil)

type sdkResponseBaseV2 struct {
	Status    *int32 `json:"status,omitempty"`
	Timestamp *int64 `json:"timestamp,omitempty"`
	Message   *struct {
		Result *string `json:"result,omitempty"`
	} `json:"message,omitempty"`
}

func (r *sdkResponseBaseV2) GetStatus() *bool {
	if r.Status != nil {
		status := *r.Status == 0
		return &status
	}

	return nil
}

func (r *sdkResponseBaseV2) GetMessage() *string {
	if r.Message != nil {
		return r.Message.Result
	}

	return nil
}

var _ sdkResponse = (*sdkResponseBaseV2)(nil)
