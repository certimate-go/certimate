package byteplusquery

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Byteplus.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/bytepluserr"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/custom"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"

	"github.com/certimate-go/certimate/pkg/sdk3rd-patched/github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/response"
)

// UnmarshalErrorHandler is a name request handler to unmarshal request errors
var UnmarshalErrorHandler = request.NamedHandler{Name: byteplusquery.UnmarshalErrorHandler.Name, Fn: UnmarshalError}

// UnmarshalError unmarshals an error response for an BYTEPLUS Query service.
func UnmarshalError(r *request.Request) {
	defer r.HTTPResponse.Body.Close()
	processUnmarshalError(unmarshalErrorInfo{
		Request: r,
	})
}

type unmarshalErrorInfo struct {
	Request  *request.Request
	Response *response.ByteplusResponse
	Body     []byte
	Err      error
}

func processUnmarshalError(info unmarshalErrorInfo) {
	var (
		body []byte
		err  error
	)
	r := info.Request
	if info.Response == nil && info.Body == nil {
		info.Response = &response.ByteplusResponse{}
		if r.DataFilled() {
			body, err = ioutil.ReadAll(r.HTTPResponse.Body)
			if err != nil {
				fmt.Printf("read byteplusbody err, %v\n", err)
				r.Error = err
				return
			}
			info.Body = body
			if err = json.Unmarshal(body, info.Response); err != nil {
				fmt.Printf("Unmarshal err, %v\n", err)
				r.Error = err
				return
			}
		} else {
			r.Error = bytepluserr.NewRequestFailure(
				bytepluserr.New("ServiceUnavailableException", "service is unavailable", nil),
				r.HTTPResponse.StatusCode,
				r.RequestID,
			)
			return
		}
	}

	if r.Config.CustomerUnmarshalError != nil {
		customerErr := r.Config.CustomerUnmarshalError(r.Context(), custom.RequestMetadata{
			ServiceName: r.ClientInfo.ServiceName,
			Version:     r.ClientInfo.APIVersion,
			Action:      r.Operation.Name,
			HttpMethod:  r.Operation.HTTPMethod,
			Region:      *r.Config.Region,
		}, *info.Response.Unwrap())
		if customerErr != nil {
			r.Error = customerErr
			return
		}
	}

	if info.Response.ResponseMetadata == nil {
		simple := response.ByteplusSimpleError{}
		if err = json.Unmarshal(info.Body, &simple); err != nil {
			fmt.Printf("Unmarshal err, %v\n", err)
			r.Error = err
			return
		}
		info.Response.ResponseMetadata = &response.ResponseMetadata{
			Error: &response.Error{
				Code:    simple.ErrorCode,
				Message: simple.Message,
			},
		}
		return
	}

	if info.Err != nil {
		r.Error = info.Err
	} else {
		r.Error = bytepluserr.NewRequestFailure(
			bytepluserr.New(info.Response.ResponseMetadata.Error.Code, info.Response.ResponseMetadata.Error.Message, nil),
			r.HTTPResponse.StatusCode,
			info.Response.ResponseMetadata.RequestId,
			r.Config.SimpleError,
		)
	}
	if reflect.TypeOf(r.Data) != reflect.TypeOf(&map[string]interface{}{}) {
		if _, ok := reflect.TypeOf(r.Data).Elem().FieldByName("Metadata"); ok {
			if info.Response.ResponseMetadata != nil {
				info.Response.ResponseMetadata.HTTPCode = r.HTTPResponse.StatusCode
			}
			r.Metadata = *info.Response.ResponseMetadata.Unwrap()
			reflect.ValueOf(r.Data).Elem().FieldByName("Metadata").Set(reflect.ValueOf(info.Response.ResponseMetadata))
		}
	}
}
