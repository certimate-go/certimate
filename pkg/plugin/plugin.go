package plugin

import (
	"context"
	"fmt"
	"log/slog"

	githubplugin "github.com/hashicorp/go-plugin"
)

const PluginName = "plugin"

type Metadata struct {
	ProviderType         string
	AccessProviderType   string
	Capability           string
	ProtocolVersion      uint32
	Category             string
	DisplayNameKey       string
	AccessDisplayNameKey string
	SupportsAccessCheck  bool
}

type ConfigSchema struct {
	AccessSchemaJSON   []byte
	ProviderSchemaJSON []byte
	I18n               map[string]map[string]string
}

type DeployPayload struct {
	CertificatePEM string
	PrivateKeyPEM  string
}

type DnsPayload struct {
	Domain  string
	Token   string
	KeyAuth string
}

type NotifyPayload struct {
	Subject string
	Message string
}

type PurgePayload struct {
	ExpirySeconds int64
}

type Request struct {
	ExecutionID        string
	Capability         string
	LogLevel           string
	AccessConfigJSON   string
	ExtendedConfigJSON string
	Deploy             *DeployPayload
	Dns                *DnsPayload
	Notify             *NotifyPayload
	Purge              *PurgePayload
}

func (r *Request) Validate() error {
	if !validCapabilities[r.Capability] {
		return fmt.Errorf("capability must be one of deploy, dns, notify, purge (got %q)", r.Capability)
	}
	var set int
	if r.Deploy != nil {
		set++
	}
	if r.Dns != nil {
		set++
	}
	if r.Notify != nil {
		set++
	}
	if r.Purge != nil {
		set++
	}
	if set != 1 {
		return fmt.Errorf("request must carry exactly one payload for capability %q", r.Capability)
	}
	switch r.Capability {
	case CapabilityDeploy:
		if r.Deploy == nil {
			return fmt.Errorf("capability %q requires a deploy payload", r.Capability)
		}
	case CapabilityDns:
		if r.Dns == nil {
			return fmt.Errorf("capability %q requires a dns payload", r.Capability)
		}
		if r.Dns.Domain == "" || r.Dns.Token == "" || r.Dns.KeyAuth == "" {
			return fmt.Errorf("dns payload requires domain, token and key auth")
		}
	case CapabilityNotify:
		if r.Notify == nil {
			return fmt.Errorf("capability %q requires a notify payload", r.Capability)
		}
	case CapabilityPurge:
		if r.Purge == nil {
			return fmt.Errorf("capability %q requires a purge payload", r.Capability)
		}
		if r.Purge.ExpirySeconds < 0 {
			return fmt.Errorf("purge expiry must be non-negative, got %d", r.Purge.ExpirySeconds)
		}
	}
	return nil
}

func (r *Request) Clone() *Request {
	if r == nil {
		return nil
	}
	dup := *r
	dup.Deploy = cloneDeployPayload(r.Deploy)
	dup.Dns = cloneDnsPayload(r.Dns)
	dup.Notify = cloneNotifyPayload(r.Notify)
	dup.Purge = clonePurgePayload(r.Purge)
	return &dup
}

func cloneDeployPayload(p *DeployPayload) *DeployPayload {
	if p == nil {
		return nil
	}
	dup := *p
	return &dup
}

func cloneDnsPayload(p *DnsPayload) *DnsPayload {
	if p == nil {
		return nil
	}
	dup := *p
	return &dup
}

func cloneNotifyPayload(p *NotifyPayload) *NotifyPayload {
	if p == nil {
		return nil
	}
	dup := *p
	return &dup
}

func clonePurgePayload(p *PurgePayload) *PurgePayload {
	if p == nil {
		return nil
	}
	dup := *p
	return &dup
}

type DeployResult struct {
	ExtendedDataJSON string
}

type DnsResult struct{}

type NotifyResult struct {
	ExtendedDataJSON string
}

type PurgeResult struct {
	Count            int
	ExtendedDataJSON string
}

type Result struct {
	Deploy *DeployResult
	Dns    *DnsResult
	Notify *NotifyResult
	Purge  *PurgeResult
}

func (res *Result) Validate(capability string) error {
	if res == nil {
		return fmt.Errorf("result must not be nil")
	}
	var set int
	if res.Deploy != nil {
		set++
	}
	if res.Dns != nil {
		set++
	}
	if res.Notify != nil {
		set++
	}
	if res.Purge != nil {
		set++
	}
	if set != 1 {
		return fmt.Errorf("result must carry exactly one result for capability %q", capability)
	}
	switch capability {
	case CapabilityDeploy:
		if res.Deploy == nil {
			return fmt.Errorf("capability %q requires a deploy result", capability)
		}
	case CapabilityDns:
		if res.Dns == nil {
			return fmt.Errorf("capability %q requires a dns result", capability)
		}
	case CapabilityNotify:
		if res.Notify == nil {
			return fmt.Errorf("capability %q requires a notify result", capability)
		}
	case CapabilityPurge:
		if res.Purge == nil {
			return fmt.Errorf("capability %q requires a purge result", capability)
		}
	default:
		return fmt.Errorf("capability must be one of deploy, dns, notify, purge (got %q)", capability)
	}
	return nil
}

type Plugin interface {
	GetMetadata(ctx context.Context) (*Metadata, error)
	GetConfigSchema(ctx context.Context) (*ConfigSchema, error)
	Process(ctx context.Context, req *Request, logger *slog.Logger) (*Result, error)
	PostProcess(ctx context.Context, req *Request, logger *slog.Logger) error
}

const (
	AccessCheckStatusPassed       = "passed"
	AccessCheckStatusFailed       = "failed"
	AccessCheckStatusInconclusive = "inconclusive"
)

const (
	AccessCheckItemStatusPassed  = "passed"
	AccessCheckItemStatusFailed  = "failed"
	AccessCheckItemStatusSkipped = "skipped"
)

var validAccessCheckStatuses = map[string]bool{
	AccessCheckStatusPassed:       true,
	AccessCheckStatusFailed:       true,
	AccessCheckStatusInconclusive: true,
}

var validAccessCheckItemStatuses = map[string]bool{
	AccessCheckItemStatusPassed:  true,
	AccessCheckItemStatusFailed:  true,
	AccessCheckItemStatusSkipped: true,
}

type AccessCheckRequest struct {
	AccessConfigJSON string
}

type AccessCheckItem struct {
	ID     string
	Status string
}

type AccessCheckResult struct {
	Status      string
	ReasonCode  string
	Items       []AccessCheckItem
	Diagnostics []string
}

func (r *AccessCheckResult) Validate() error {
	if r == nil {
		return fmt.Errorf("access check result must not be nil")
	}
	if !validAccessCheckStatuses[r.Status] {
		return fmt.Errorf("access check status must be one of passed, failed, inconclusive (got %q)", r.Status)
	}
	for _, item := range r.Items {
		if item.ID == "" {
			return fmt.Errorf("access check item must carry an id")
		}
		if !validAccessCheckItemStatuses[item.Status] {
			return fmt.Errorf("access check item status must be one of passed, failed, skipped (got %q)", item.Status)
		}
	}
	return nil
}

type AccessChecker interface {
	CheckAccess(ctx context.Context, req *AccessCheckRequest, logger *slog.Logger) (*AccessCheckResult, error)
}

type NoopFinalizer struct{}

func (NoopFinalizer) PostProcess(context.Context, *Request, *slog.Logger) error {
	return nil
}

type PluginSet = githubplugin.PluginSet

func ServeConfig(impl Plugin) *githubplugin.ServeConfig {
	return &githubplugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         PluginSet{PluginName: &GRPCPlugin{Impl: impl}},
		GRPCServer:      githubplugin.DefaultGRPCServer,
	}
}
