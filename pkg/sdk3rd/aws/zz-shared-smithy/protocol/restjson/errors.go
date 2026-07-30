package restjson

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/smithy-go"
)

func GetAPIError(data []byte) (smithy.APIError, error) {
	return GetAPIErrorWithRawResponse(data, nil)
}

func GetAPIErrorWithRawResponse(data []byte, rawResp *http.Response) (smithy.APIError, error) {
	errorMessage, err := parseAPIErrorInfo(data)
	if err != nil {
		return nil, err
	}

	apiError := &smithy.GenericAPIError{
		Message: errorMessage,
	}
	if rawResp != nil {
		apiError.Code = rawResp.Header.Get("X-Amzn-ErrorType")

		switch rawStatus := rawResp.StatusCode; {
		case rawStatus >= http.StatusBadRequest && rawStatus < http.StatusInternalServerError:
			apiError.Fault = smithy.FaultClient
		case rawStatus >= http.StatusInternalServerError:
			apiError.Fault = smithy.FaultServer
		}
	}

	return apiError, nil
}

func parseAPIErrorInfo(data []byte) (errorMessage string, err error) {
	var errInfo struct {
		Message string
	}

	err = json.Unmarshal(data, &errInfo)
	if err != nil {
		if err == io.EOF {
			return errorMessage, nil
		}
		return errorMessage, err
	}

	if len(errInfo.Message) != 0 {
		errorMessage = errInfo.Message
	}

	return errorMessage, nil
}
