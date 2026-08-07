package tos

type sdkResponse interface{}

type sdkResponseBase struct{}

var _ sdkResponse = (*sdkResponseBase)(nil)
