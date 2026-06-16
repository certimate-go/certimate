package nginxproxymanager

import (
	"encoding/json"
	"fmt"
)

type sdkResponse interface {
	GetError() string
}

type sdkResponseBase struct {
	Error json.RawMessage `json:"error"`
}

func (r *sdkResponseBase) GetError() string {
	if len(r.Error) == 0 {
		return ""
	}

	var errStr string
	if err := json.Unmarshal(r.Error, &errStr); err == nil {
		return errStr
	}

	type errObjType struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	var errObj errObjType
	if err := json.Unmarshal(r.Error, &errObj); err == nil && errObj.Message != "" {
		if errObj.Code != 0 {
			return fmt.Sprintf("%d %s", errObj.Code, errObj.Message)
		}
		return errObj.Message
	}

	var errMap map[string]any
	if err := json.Unmarshal(r.Error, &errMap); err == nil {
		if message, ok := errMap["message"].(string); ok {
			return message
		}
	}

	return ""
}

var _ sdkResponse = (*sdkResponseBase)(nil)
