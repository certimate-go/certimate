package awsclb

import (
	cmgrimplacm "github.com/certimate-go/certimate/pkg/core/certmgr/providers/aws-acm"
)

const (
	AUTH_METHOD_ACCESSKEY = cmgrimplacm.AUTH_METHOD_ACCESSKEY
	AUTH_METHOD_IMDS      = cmgrimplacm.AUTH_METHOD_IMDS
)

const (
	CERTIFICATE_SOURCE_ACM = "ACM"
	CERTIFICATE_SOURCE_IAM = "IAM"
)
