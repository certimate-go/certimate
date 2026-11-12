package byteplusquery

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Byteplus.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/bytepluserr"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusutil"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/special"

	"github.com/certimate-go/certimate/pkg/sdk3rd-patched/github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/response"
)

// UnmarshalHandler is a named request handler for unmarshaling byteplusquery protocol requests
var UnmarshalHandler = request.NamedHandler{Name: byteplusquery.UnmarshalHandler.Name, Fn: Unmarshal}

// UnmarshalMetaHandler is a named request handler for unmarshaling byteplusquery protocol request metadata
var UnmarshalMetaHandler = request.NamedHandler{Name: byteplusquery.UnmarshalMetaHandler.Name, Fn: UnmarshalMeta}

// Unmarshal unmarshals a response for an BYTEPLUS Query service.
func Unmarshal(r *request.Request) {
	defer r.HTTPResponse.Body.Close()
	if r.DataFilled() {
		body, err := ioutil.ReadAll(r.HTTPResponse.Body)
		if err != nil {
			fmt.Printf("read byteplusbody err, %v\n", err)
			r.Error = err
			return
		}

		var forceJsonNumberDecoder bool

		if r.Config.ForceJsonNumberDecode != nil {
			forceJsonNumberDecoder = r.Config.ForceJsonNumberDecode(r.Context(), r.MergeRequestInfo())
		}

		if reflect.TypeOf(r.Data) == reflect.TypeOf(&map[string]interface{}{}) {
			// 如果使用map返回 发现精度丢失了 请设置强制JsonNumber 注意返回的整型会动float64->int64
			if err = json.Unmarshal(body, &r.Data); err != nil || forceJsonNumberDecoder {
				// try next
				decoder := json.NewDecoder(bytes.NewReader(body))
				decoder.UseNumber()
				if err = decoder.Decode(&r.Data); err != nil {
					fmt.Printf("Unmarshal err, %v\n", err)
					r.Error = err
					return
				}
			}
			var info interface{}

			ptr := r.Data.(*map[string]interface{})
			info, err = byteplusutil.ObtainSdkValue("ResponseMetadata.Error.Code", *ptr)
			if err != nil {
				r.Error = err
				return
			}
			if info != nil {
				if processBodyError(r, &response.ByteplusResponse{}, body, forceJsonNumberDecoder) {
					return
				}
			}

		} else {
			byteplusResponse := response.ByteplusResponse{}
			if processBodyError(r, &byteplusResponse, body, forceJsonNumberDecoder) {
				return
			}

			if _, ok := reflect.TypeOf(r.Data).Elem().FieldByName("Metadata"); ok {
				if byteplusResponse.ResponseMetadata != nil {
					byteplusResponse.ResponseMetadata.HTTPCode = r.HTTPResponse.StatusCode
				}
				r.Metadata = *byteplusResponse.ResponseMetadata.Unwrap()
				reflect.ValueOf(r.Data).Elem().FieldByName("Metadata").Set(reflect.ValueOf(byteplusResponse.ResponseMetadata))
			}

			var (
				b      []byte
				source interface{}
			)

			if r.Config.CustomerUnmarshalData != nil {
				source = r.Config.CustomerUnmarshalData(r.Context(), r.MergeRequestInfo(), *byteplusResponse.Unwrap())
			} else {
				if sp, ok := special.ResponseSpecialMapping()[r.ClientInfo.ServiceName]; ok {
					source = sp(*byteplusResponse.Unwrap(), r.Data)
				} else {
					source = byteplusResponse.Result
				}
			}

			if b, err = json.Marshal(source); err != nil {
				fmt.Printf("Unmarshal err, %v\n", err)
				r.Error = err
				return
			}
			if err = json.Unmarshal(b, &r.Data); err != nil || forceJsonNumberDecoder {
				decoder := json.NewDecoder(bytes.NewReader(b))
				decoder.UseNumber()
				if err = decoder.Decode(&r.Data); err != nil {
					fmt.Printf("Unmarshal err, %v\n", err)
					r.Error = err
					return
				}
			}
		}

	}
}

// UnmarshalMeta unmarshals header response values for an BYTEPLUS Query service.
func UnmarshalMeta(r *request.Request) {
}

func processBodyError(r *request.Request, byteplusResponse *response.ByteplusResponse, body []byte, forceJsonNumberDecoder bool) bool {
	// 防止精度问题 第一次转换 无视 保持原body内容不会失去精度
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&byteplusResponse); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		r.Error = err
		return true
	}

	if byteplusResponse.ResponseMetadata.Error != nil && byteplusResponse.ResponseMetadata.Error.Code != "" {
		r.Error = bytepluserr.NewRequestFailure(
			bytepluserr.New(byteplusResponse.ResponseMetadata.Error.Code, byteplusResponse.ResponseMetadata.Error.Message, nil),
			r.HTTPResponse.StatusCode,
			byteplusResponse.ResponseMetadata.RequestId,
		)
		processUnmarshalError(unmarshalErrorInfo{
			Request:  r,
			Response: byteplusResponse,
			Body:     body,
			Err:      r.Error,
		})
		return true
	}
	return false
}
