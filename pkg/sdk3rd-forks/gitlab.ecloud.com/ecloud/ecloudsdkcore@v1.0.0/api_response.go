package ecloudsdkcore

import (
	"net/http"
)

type ReturnState string

const (
	OK        ReturnState = "OK"
	ERROR     ReturnState = "ERROR"
	EXCEPTION ReturnState = "EXCEPTION"
	ALARM     ReturnState = "ALARM"
	FORBIDDEN ReturnState = "FORBIDDEN"
)

type APIResponse struct {
	*http.Response `json:"-"`
	Message        string `json:"message,omitempty"`
	// Operation is the name of the swagger operation.
	Operation string `json:"operation,omitempty"`
	// RequestURL is the request URL. This value is always available, even if the
	// embedded *http.Response is nil.
	RequestURL string `json:"url,omitempty"`
	// Method is the HTTP method used for the request.  This value is always
	// available, even if the embedded *http.Response is nil.
	Method string `json:"method,omitempty"`
	// Payload holds the contents of the response body (which may be nil or empty).
	// This is provided here as the raw response.Body() reader will have already
	// been drained.
	Payload []byte `json:"-"`
}

type APIPlatformResponse struct {
	RequestId    string      `json:"requestId,omitempty"`
	State        ReturnState `json:"state,omitempty"`
	Body         interface{} `json:"body,omitempty"`
	ErrorCode    string      `json:"errorCode,omitempty"`
	ErrorParams  []string    `json:"errorParams,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
}

type APIPlatformResponseBody struct {
	// TimeConsuming   int64                  `json:"timeConsuming,omitempty"`
	ResponseBody    string                 `json:"responseBody,omitempty"`
	RequestHeader   map[string]interface{} `json:"requestHeader,omitempty"`
	ResponseHeader  map[string]interface{} `json:"responseHeader,omitempty"`
	ResponseMessage string                 `json:"responseMessage,omitempty"`
	StatusCode      int                    `json:"statusCode,omitempty"`
	HttpMethod      string                 `json:"httpMethod,omitempty"`
	RequestUrl      string                 `json:"requestUrl,omitempty"`
}

func NewAPIResponse(r *http.Response) *APIResponse {
	response := &APIResponse{Response: r}
	return response
}

func NewAPIResponseWithError(errorMessage string) *APIResponse {
	response := &APIResponse{Message: errorMessage}
	return response
}
