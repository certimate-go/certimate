package awslightsail

import (
	route53 "github.com/certimate-go/certimate/pkg/core/certifier/challengers/dns01/aws-route53"
)

const (
	AUTH_METHOD_ACCESSKEY = route53.AUTH_METHOD_ACCESSKEY
	AUTH_METHOD_IMDS      = route53.AUTH_METHOD_IMDS
)
