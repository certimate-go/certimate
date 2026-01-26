package internal

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/samber/lo"

	xinnetsdk "github.com/certimate-go/certimate/pkg/sdk3rd/xinnet"
)

const (
	envNamespace = "XINNET_"

	EnvAgentId   = envNamespace + "AGENT_ID"
	EnvAppSecret = envNamespace + "APP_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AgentID   string
	AppSecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *xinnetsdk.Client

	recordIDs   map[string]*int64 // Key: ChallengeToken; Value: RecordID
	recordIDsMu sync.Mutex
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAgentId, EnvAppSecret)
	if err != nil {
		return nil, fmt.Errorf("xinnet: %w", err)
	}

	config := NewDefaultConfig()
	config.AgentID = values[EnvAgentId]
	config.AppSecret = values[EnvAppSecret]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("xinnet: the configuration of the DNS provider is nil")
	}

	client, err := xinnetsdk.NewClient(config.AgentID, config.AppSecret)
	if err != nil {
		return nil, fmt.Errorf("xinnet: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		recordIDs:   make(map[string]*int64),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("xinnet: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://apidoc.xin.cn/doc-7283900
	request := &xinnetsdk.DnsCreateRequest{
		DomainName: lo.ToPtr(dns01.UnFqdn(authZone)),
		RecordName: lo.ToPtr(dns01.UnFqdn(info.EffectiveFQDN)),
		Type:       lo.ToPtr("TXT"),
		Value:      lo.ToPtr(info.Value),
		Line:       lo.ToPtr("默认"),
		Ttl:        lo.ToPtr(int32(d.config.TTL)),
	}
	response, err := d.client.DnsCreate(request)
	if err != nil {
		return fmt.Errorf("xinnet: error when create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = response.Data
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("xinnet: could not find zone for domain %q: %w", domain, err)
	}

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("xinnet: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	// REF: https://apidoc.xin.cn/doc-7283901
	request := &xinnetsdk.DnsDeleteRequest{
		DomainName: lo.ToPtr(dns01.UnFqdn(authZone)),
		RecordId:   recordID,
	}
	if _, err := d.client.DnsDelete(request); err != nil {
		return fmt.Errorf("xinnet: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
