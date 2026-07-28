package restxml

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/aws/smithy-go"
)

func GetAPIError(data []byte) (smithy.APIError, error) {
	return GetAPIErrorWithRawResponse(data, nil)
}

func GetAPIErrorWithRawResponse(data []byte, rawResp *http.Response) (smithy.APIError, error) {
	errorType, errorMessage, err := parseAPIErrorInfo(data)
	if err != nil {
		return nil, err
	}

	apiError := &smithy.GenericAPIError{
		Code:    errorType,
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
		Code    string `xml:"Error>Code"`
		Message string `xml:"Error>Message"`
	}

	err = xml.Unmarshal(data, &errInfo)
	if err != nil {
		if err == io.EOF {
			return errorCode, errorMessage, nil
		}
		return errorCode, errorMessage, err
	}

	if len(errInfo.Code) != 0 {
		errorCode = errInfo.Code
	}

	if len(errInfo.Message) != 0 {
		errorMessage = errInfo.Message
	}

	return errorCode, errorMessage, nil
}
