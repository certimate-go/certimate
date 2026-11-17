package teomakers

import (
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type (
	OwnershipVerification = teo.OwnershipVerification
	DnsVerification       = teo.DnsVerification
	FileVerification      = teo.FileVerification
	NsVerification        = teo.NsVerification
)

type PagesZoneCustomDomain struct {
	Type                  *string                `json:"Type,omitempty"`
	Domain                *string                `json:"Domain,omitempty"`
	ForceRedirectHTTPS    *string                `json:"ForceRedirectHTTPS,omitempty"`
	RedirectStatusCode    *int32                 `json:"RedirectStatusCode,omitempty"`
	CurrentCname          *string                `json:"CurrentCname,omitempty"`
	MainDomain            *string                `json:"MainDomain,omitempty"`
	Status                *string                `json:"Status,omitempty"`
	OwnershipVerification *OwnershipVerification `json:"OwnershipVerification,omitempty"`
	Cname                 *string                `json:"Cname,omitempty"`
	Area                  *string                `json:"Area,omitempty"`
	ZoneId                *string                `json:"ZoneId,omitempty"`
}
