package query

import (
	"net/http"

	"github.com/aws/smithy-go"

	"github.com/certimate-go/certimate/pkg/sdk3rd/aws/zz-shared-smithy/protocol/restxml"
)

func GetAPIError(data []byte) (smithy.APIError, error) {
	return GetAPIErrorWithRawResponse(data, nil)
}

func GetAPIErrorWithRawResponse(data []byte, rawResp *http.Response) (smithy.APIError, error) {
	return restxml.GetAPIErrorWithRawResponse(data, rawResp)
}
