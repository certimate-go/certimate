package dtos

type AccessTestSource string

const (
	AccessTestSourceStored AccessTestSource = "stored"
	AccessTestSourceDraft  AccessTestSource = "draft"
)

type AccessTestStatus string

const (
	AccessTestStatusPassed       AccessTestStatus = "passed"
	AccessTestStatusFailed       AccessTestStatus = "failed"
	AccessTestStatusInconclusive AccessTestStatus = "inconclusive"
	AccessTestStatusUnsupported  AccessTestStatus = "unsupported"
	AccessTestStatusUnavailable  AccessTestStatus = "unavailable"
)

type AccessTestItemStatus string

const (
	AccessTestItemStatusPassed  AccessTestItemStatus = "passed"
	AccessTestItemStatusFailed  AccessTestItemStatus = "failed"
	AccessTestItemStatusSkipped AccessTestItemStatus = "skipped"
)

type AccessTestReasonCode string

const (
	AccessTestReasonCodeAuthRejected       AccessTestReasonCode = "auth_rejected"
	AccessTestReasonCodeConnectFailed      AccessTestReasonCode = "connect_failed"
	AccessTestReasonCodeTlsFailed          AccessTestReasonCode = "tls_failed"
	AccessTestReasonCodeTimeout            AccessTestReasonCode = "timeout"
	AccessTestReasonCodeCanceled           AccessTestReasonCode = "canceled"
	AccessTestReasonCodeProbeForbidden     AccessTestReasonCode = "probe_forbidden"
	AccessTestReasonCodeRemoteError        AccessTestReasonCode = "remote_error"
	AccessTestReasonCodeNoChecker          AccessTestReasonCode = "no_checker"
	AccessTestReasonCodePluginMissing      AccessTestReasonCode = "plugin_missing"
	AccessTestReasonCodePluginUnavailable  AccessTestReasonCode = "plugin_unavailable"
	AccessTestReasonCodePluginProtocolOld  AccessTestReasonCode = "plugin_protocol_old"
	AccessTestReasonCodeConfigInvalid      AccessTestReasonCode = "config_invalid"
	AccessTestReasonCodeInternalError      AccessTestReasonCode = "internal_error"
	AccessTestReasonCodeConcurrencyLimited AccessTestReasonCode = "concurrency_limited"
	AccessTestReasonCodeAccessNotFound     AccessTestReasonCode = "access_not_found"
)

type AccessTestCapabilitiesResp struct {
	Provider          string   `json:"provider"`
	Supported         bool     `json:"supported"`
	Source            string   `json:"source"`
	OwnerProviderType string   `json:"ownerProviderType"`
	Scope             []string `json:"scope"`
	ReasonCode        string   `json:"reasonCode"`
}

type AccessTestReq struct {
	Source   AccessTestSource `json:"source"`
	AccessId string           `json:"accessId,omitempty"`
	Provider string           `json:"provider,omitempty"`
	Config   map[string]any   `json:"config,omitempty"`
}

type AccessTestResp struct {
	Status      AccessTestStatus     `json:"status"`
	ReasonCode  string               `json:"reasonCode"`
	Items       []AccessTestItemResp `json:"items"`
	Diagnostics []string             `json:"diagnostics"`
}

type AccessTestItemResp struct {
	Id     string               `json:"id"`
	Status AccessTestItemStatus `json:"status"`
}
