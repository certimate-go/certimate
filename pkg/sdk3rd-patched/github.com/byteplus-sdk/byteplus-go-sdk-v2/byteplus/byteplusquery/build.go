package byteplusquery

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Byteplus.

import (
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/byteplusquery"
	"github.com/byteplus-sdk/byteplus-go-sdk-v2/byteplus/request"
)

// BuildHandler is a named request handler for building byteplusquery protocol requests
var BuildHandler = request.NamedHandler{Name: byteplusquery.BuildHandler.Name, Fn: Build}

// Build builds a request for a byteplus Query service.
func Build(r *request.Request) {
	byteplusquery.Build(r)
}
