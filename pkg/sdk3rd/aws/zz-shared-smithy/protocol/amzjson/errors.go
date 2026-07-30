package amzjson

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/restjson"
	"github.com/aws/smithy-go"
)

func GetAPIInfo(data []byte) (smithy.APIError, error) {
	return GetAPIErrorWithRawResponse(data, nil)
}

func GetAPIErrorWithRawResponse(data []byte, rawResp *http.Response) (smithy.APIError, error) {
	errorCode, errorMessage, err := parseAPIErrorInfo(data)
	if err != nil {
		return nil, err
	}

	apiError := &smithy.GenericAPIError{
		Code:    errorCode,
		Message: errorMessage,
	}
	if rawResp != nil {
		switch rawStatus := rawResp.StatusCode; {
		case rawStatus >= http.StatusBadRequest && rawStatus < http.StatusInternalServerError:
			apiError.Fault = smithy.FaultClient
		case rawStatus >= http.StatusInternalServerError:
			apiError.Fault = smithy.FaultServer
		}
	}

	return apiError, nil
}

func parseAPIErrorInfo(data []byte) (errorCode string, errorMessage string, err error) {
	var errInfo struct {
		Code    string
		Type    string `json:"__type"`
		Message string
	}

	err = json.Unmarshal(data, &errInfo)
	if err != nil {
		if err == io.EOF {
			return errorCode, errorMessage, nil
		}
		return errorCode, errorMessage, err
	}

	if len(errInfo.Code) != 0 {
		errorCode = errInfo.Code
	} else if len(errInfo.Type) != 0 {
		errorCode = errInfo.Type
	}

	if len(errInfo.Message) != 0 {
		errorMessage = errInfo.Message
	}

	if len(errorCode) != 0 {
		errorCode = restjson.SanitizeErrorCode(errorCode)
	}

	return errorCode, errorMessage, nil
}
