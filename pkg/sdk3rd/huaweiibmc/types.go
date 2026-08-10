package huaweiibmc

import (
	"encoding/json"
	"fmt"
)

type sdkResponse interface {
	GetAPIError() error
}

type sdkResponseBase struct {
	ODataContext string       `json:"@odata.context,omitempty"`
	Error        *sdkAPIError `json:"error,omitempty"`
}

type sdkAPIError struct {
	Code                string          `json:"code"`
	Message             string          `json:"message"`
	MessageExtendedInfo json.RawMessage `json:"@Message.ExtendedInfo,omitempty"`
}

func (e sdkAPIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (r *sdkResponseBase) GetAPIError() error {
	if r.Error != nil {
		return *r.Error
	}
	return nil
}

var _ sdkResponse = (*sdkResponseBase)(nil)
