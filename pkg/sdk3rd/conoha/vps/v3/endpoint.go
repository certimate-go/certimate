package v3

import "fmt"

const region = "c3j1"

var (
	identityBaseURL = fmt.Sprintf("https://identity.%s.conoha.io", region)
	dnsBaseURL      = fmt.Sprintf("https://dns-service.%s.conoha.io", region)
)
