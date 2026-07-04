package eomakers

type DescribePagesZoneCustomDomains struct {
	Type                  string `json:"Type"`
	Domain                string `json:"Domain"`
	ForceRedirectHTTPS    string `json:"ForceRedirectHTTPS"`
	RedirectStatusCode    int    `json:"RedirectStatusCode"`
	CurrentCname          string `json:"CurrentCname"`
	MainDomain            string `json:"MainDomain"`
	Status                string `json:"Status"`
	OwnershipVerification struct {
		DnsVerification *struct {
			RecordType  string `json:"RecordType"`
			RecordValue string `json:"RecordValue"`
			Subdomain   string `json:"Subdomain"`
		} `json:"DnsVerification,omitempty"`
		FileVerification *struct {
			Content string `json:"Content"`
			Path    string `json:"Path"`
		} `json:"FileVerification,omitempty"`
		NsVerification any `json:"NsVerification"`
	} `json:"OwnershipVerification"`
	Cname  string `json:"Cname"`
	Area   string `json:"Area"`
	ZoneId string `json:"ZoneId"`
}

type DescribePagesZoneCustomDomainsResp struct {
	Code int `json:"Code"`
	Data struct {
		Response struct {
			RequestId    string                           `json:"RequestId"`
			PagesDomains []DescribePagesZoneCustomDomains `json:"PagesDomains"`
			TotalCount   int                              `json:"TotalCount"`
		} `json:"Response"`
	} `json:"Data"`
}

type DescribePagesZoneCustomDomainsReq struct {
	Action    string `json:"Action"`
	ProjectId string `json:"ProjectId"`
}

type DescribeMakersZoneCustomDomains = DescribePagesZoneCustomDomains

type DescribeMakersZoneCustomDomainsResp = DescribePagesZoneCustomDomainsResp

type DescribeMakersZoneCustomDomainsReq = DescribePagesZoneCustomDomainsReq
