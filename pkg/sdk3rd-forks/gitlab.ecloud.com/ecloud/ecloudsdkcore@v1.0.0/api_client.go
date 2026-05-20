package ecloudsdkcore

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
)

var (
	jsonCheck = regexp.MustCompile("(?i:(?:application|text)/json)")
	xmlCheck  = regexp.MustCompile("(?i:(?:application|text)/xml)")
)

// APIClient manages communication
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *Configuration
	common service
}

type service struct {
	client *APIClient
}

type HttpRequestPosition string

const (
	BODY   HttpRequestPosition = "Body"
	QUERY  HttpRequestPosition = "Query"
	PATH   HttpRequestPosition = "Path"
	HEADER HttpRequestPosition = "Header"
)

const (
	SdkPortalUrl        = "/op-apim-portal/apim/request/sdk"
	SdkPortalGatewayUrl = "/api/query/openapi/apim/request/sdk"
)

// NewAPIClient creates a new API client.
func NewAPIClient() *APIClient {
	cfg := NewConfiguration()
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c
	return c
}

// atoi string to int
func atoi(in string) (int, error) {
	return strconv.Atoi(in)
}

// selectHeaderContentType select a content type from the available list.
func selectHeaderContentType(contentTypes []string) string {
	if len(contentTypes) == 0 {
		return ""
	}
	if contains(contentTypes, "application/json") {
		return "application/json"
	}
	return contentTypes[0]
}

// selectHeaderAccept join all accept types and return
func selectHeaderAccept(accepts []string) string {
	if len(accepts) == 0 {
		return ""
	}

	if contains(accepts, "application/json") {
		return "application/json"
	}

	return strings.Join(accepts, ",")
}

// contains is a case insenstive match, finding needle in a haystack
func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if strings.ToLower(a) == strings.ToLower(needle) {
			return true
		}
	}
	return false
}

// Verify optional parameters are of the correct type.
func typeCheckParameter(obj interface{}, expected string, name string) error {
	if obj == nil {
		return nil
	}
	if reflect.TypeOf(obj).String() != expected {
		return fmt.Errorf("Expected %s to be of type %s but received %s.", name, expected, reflect.TypeOf(obj).String())
	}
	return nil
}

// parameterToString convert interface{} parameters to string, using a delimiter if format is provided.
func parameterToString(obj interface{}, collectionFormat string, request HttpRequest) (*http.Request, string) {
	var delimiter string

	switch collectionFormat {
	case "pipes":
		delimiter = "|"
	case "ssv":
		delimiter = " "
	case "tsv":
		delimiter = "\t"
	case "csv":
		delimiter = ","
	}

	if reflect.TypeOf(obj).Kind() == reflect.Slice {
		return nil, strings.Trim(strings.Replace(fmt.Sprint(obj), " ", delimiter, -1), "[]")
	}

	return nil, fmt.Sprintf("%v", obj)
}

// Excute entry for http call
func (c *APIClient) Excute(httpRequest *HttpRequest, config *config.Config, returnType interface{}) (*http.Response, error) {
	httpRequest = buildHttpRequest(httpRequest, config)
	request := buildCall(httpRequest)
	httpResponse, err := c.callAPI(request)
	if err != nil || httpResponse == nil {
		return nil, err
	}

	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if err != nil {
		return httpResponse, err
	}

	if httpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = c.decode(&returnType, responseBody, httpResponse.Header.Get("Content-Type"))
		if err != nil {
			return httpResponse, fmt.Errorf("%w, response body is: %s", err, string(responseBody))
		}
		return httpResponse, nil
	}

	if httpResponse.StatusCode >= 300 {
		newErr := GenericResponseError{
			body:  responseBody,
			error: httpResponse.Status,
		}
		return httpResponse, newErr
	}
	return httpResponse, err
}

// callAPI do the request.
func (c *APIClient) callAPI(request *http.Request) (*http.Response, error) {
	return c.cfg.HTTPClient.Do(request)
}

// ChangeBasePath Change base path to allow switching to mocks
func (c *APIClient) ChangeBasePath(path string) {
	c.cfg.BasePath = path
}

// buildHttpRequest build the request
func buildHttpRequest(httpRequest *HttpRequest, config *config.Config) *HttpRequest {
	openApiRequest := &OpenApiRequest{
		AccessKey:  config.AccessKey,
		SecretKey:  config.SecretKey,
		PoolId:     config.PoolId,
		Api:        httpRequest.Action,
		Product:    httpRequest.Product,
		Version:    httpRequest.Version,
		SdkVersion: httpRequest.SdkVersion,
		Language:   "Golang",
	}
	if httpRequest.Body != nil {
		reqType := reflect.TypeOf(httpRequest.Body)
		if reqType.Kind() == reflect.Ptr {
			reqType = reqType.Elem()
		}
		v := reflect.ValueOf(httpRequest.Body)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		flag := false
		for i := 0; i < reqType.NumField(); i++ {
			fieldType := reqType.Field(i)
			value := v.FieldByName(fieldType.Name)
			if value.Kind() == reflect.Ptr {
				if value.IsNil() {
					continue
				}
				value = value.Elem()

			}
			propertyType := fieldType.Type
			if propertyType.Kind() == reflect.Ptr {
				propertyType = propertyType.Elem()
			}

			_, flag = propertyType.FieldByName(string(BODY))
			if flag {
				openApiRequest.BodyParameter = value.Interface()
				continue
			}
			_, flag = propertyType.FieldByName(string(HEADER))
			if flag {
				openApiRequest.HeaderParameter = structToMap(value.Interface())
				continue
			}
			_, flag = propertyType.FieldByName(string(QUERY))
			if flag {
				openApiRequest.QueryParameter = structToMap(value.Interface())
				continue
			}
			_, flag = propertyType.FieldByName(string(PATH))
			if flag {
				openApiRequest.PathParameter = structToMap(value.Interface())
				continue
			}
		}
	}
	headers := make(map[string]interface{})
	if httpRequest.HeaderParams != nil {
		if openApiRequest.HeaderParameter == nil {
			headers = httpRequest.HeaderParams
		} else {
			headers = mergeMap(openApiRequest.HeaderParameter, httpRequest.HeaderParams)
		}
		openApiRequest.HeaderParameter = headers
	}
	httpRequest.Body = openApiRequest
	return httpRequest
}

// mergeMap merge the two map results
func mergeMap(mObj ...map[string]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	for _, m := range mObj {
		for k, v := range m {
			newMap[k] = v
		}
	}
	return newMap
}

// structToMap struct convert to map
func structToMap(value interface{}) map[string]interface{} {
	data, _ := json.Marshal(value)
	result := make(map[string]interface{})
	json.Unmarshal(data, &result)
	return result
}

func buildCall(httpRequest *HttpRequest) (request *http.Request) {
	url := ""
	if len(httpRequest.Url) > 0 {
		url = httpRequest.Url + SdkPortalUrl
	} else {
		url = httpRequest.DefaultUrl + SdkPortalGatewayUrl
	}
	request, _ = prepareRequest(url, "POST", httpRequest.Body)
	return request
}

// prepareRequest build the request
func prepareRequest(path string, method string,
	postBody interface{},
) (httpRequest *http.Request, err error) {
	var body *bytes.Buffer

	// Detect postBody type and post.
	if postBody != nil {
		contentType := detectContentType(postBody)
		body, err = setBody(postBody, contentType)
		if err != nil {
			return nil, err
		}
	}

	// Setup path and query parameters
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Generate a new request
	if body != nil {
		httpRequest, err = http.NewRequest(method, url.String(), body)
	} else {
		httpRequest, err = http.NewRequest(method, url.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	// add default header parameters
	httpRequest.Header.Add("Content-Type", "application/json")
	return httpRequest, nil
}

func (c *APIClient) decode(v interface{}, b []byte, contentType string) (err error) {
	if strings.Contains(contentType, "application/xml") {
		if err = xml.Unmarshal(b, v); err != nil {
			return err
		}
		return nil
	} else if strings.Contains(contentType, "application/json") {
		platformResponse := &APIPlatformResponse{}
		if err = json.Unmarshal(b, platformResponse); err != nil {
			newErr := GenericResponseError{
				body:  b,
				error: err.Error(),
			}
			return newErr
		}
		platformResponseBodyBytes, _ := json.Marshal(platformResponse.Body)
		platformResponseBody := &APIPlatformResponseBody{}
		if err = json.Unmarshal(platformResponseBodyBytes, platformResponseBody); err != nil {
			return err
		}
		/*
			找到两层指针指向的元素
		*/
		value := reflect.ValueOf(v).Elem().Elem()

		if !value.IsNil() {
			structValue := value.Elem()
			if structValue.NumField() == 1 && structValue.Field(0).Kind() == reflect.String {
				n := len(platformResponseBody.ResponseBody)
				structValue.Field(0).SetString(platformResponseBody.ResponseBody[1 : n-1])
				return nil
			}
		}

		if err = json.Unmarshal([]byte(platformResponseBody.ResponseBody), v); err != nil {
			return err
		}
		return nil
	}
	return errors.New("undefined response type")
}

// Add a file to the multipart request
func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return err
}

// Prevent trying to import "fmt"
func reportError(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

// Set request body from an interface{}
func setBody(body interface{}, contentType string) (bodyBuf *bytes.Buffer, err error) {
	if bodyBuf == nil {
		bodyBuf = &bytes.Buffer{}
	}
	if reader, ok := body.(io.Reader); ok {
		_, err = bodyBuf.ReadFrom(reader)
	} else if b, ok := body.([]byte); ok {
		_, err = bodyBuf.Write(b)
	} else if s, ok := body.(string); ok {
		_, err = bodyBuf.WriteString(s)
	} else if s, ok := body.(*string); ok {
		_, err = bodyBuf.WriteString(*s)
	} else if jsonCheck.MatchString(contentType) {
		err = json.NewEncoder(bodyBuf).Encode(body)
	} else if xmlCheck.MatchString(contentType) {
		xml.NewEncoder(bodyBuf).Encode(body)
	}

	if err != nil {
		return nil, err
	}

	if bodyBuf.Len() == 0 {
		err = fmt.Errorf("Invalid body type %s\n", contentType)
		return nil, err
	}
	return bodyBuf, nil
}

// detectContentType method is used to figure out `Request.Body` content type for request header
func detectContentType(body interface{}) string {
	contentType := "text/plain; charset=utf-8"
	kind := reflect.TypeOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Ptr:
		contentType = "application/json; charset=utf-8"
	case reflect.String:
		contentType = "text/plain; charset=utf-8"
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = "application/json; charset=utf-8"
		}
	}

	return contentType
}

type cacheControl map[string]string

func parseCacheControl(headers http.Header) cacheControl {
	cc := cacheControl{}
	ccHeader := headers.Get("Cache-Control")
	for _, part := range strings.Split(ccHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyval := strings.Split(part, "=")
			cc[strings.Trim(keyval[0], " ")] = strings.Trim(keyval[1], ",")
		} else {
			cc[part] = ""
		}
	}
	return cc
}

// CacheExpires helper function to determine remaining time before repeating a request.
func CacheExpires(r *http.Response) time.Time {
	// Figure out when the cache expires.
	var expires time.Time
	now, err := time.Parse(time.RFC1123, r.Header.Get("date"))
	if err != nil {
		return time.Now()
	}
	respCacheControl := parseCacheControl(r.Header)

	if maxAge, ok := respCacheControl["max-age"]; ok {
		lifetime, err := time.ParseDuration(maxAge + "s")
		if err != nil {
			expires = now
		}
		expires = now.Add(lifetime)
	} else {
		expiresHeader := r.Header.Get("Expires")
		if expiresHeader != "" {
			expires, err = time.Parse(time.RFC1123, expiresHeader)
			if err != nil {
				expires = now
			}
		}
	}
	return expires
}

func strlen(s string) int {
	return utf8.RuneCountInString(s)
}

// GenericResponseError Provides access to the body, error and model on returned errors.
type GenericResponseError struct {
	body  []byte
	error string
	model interface{}
}

// Error returns non-empty string if there was an error.
func (e GenericResponseError) Error() string {
	return e.error
}

// Body returns the raw bytes of the response
func (e GenericResponseError) Body() []byte {
	return e.body
}

// Model returns the unpacked model of the error
func (e GenericResponseError) Model() interface{} {
	return e.model
}
