package internal

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/samber/lo"

	dynadotsdk "github.com/certimate-go/certimate/pkg/sdk3rd/dynadot"
)

const (
	envNamespace = "DYNADOT_"

	EnvAPIKey    = envNamespace + "API_KEY"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	APIKey    string
	APISecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *dynadotsdk.Client

	recordCache   map[string]dnsRecordCacheEntry // Key: ChallengeToken
	recordCacheMu sync.Mutex
}

type dnsRecordCacheEntry struct {
	Zone    string
	SubHost string
	Value   string
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("dynadot: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("dynadot: the configuration of the DNS provider is nil")
	}

	client, err := dynadotsdk.NewClient(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("dynadot: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:        config,
		client:        client,
		recordCache:   make(map[string]dnsRecordCacheEntry),
		recordCacheMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynadot: could not find zone for domain %q: %w", domain, err)
	}

	zone := dns01.UnFqdn(authZone)
	subHost, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dynadot: %w", err)
	}

	// REF: https://www.dynadot.com/domain/api-document (set_dns)
	request := &dynadotsdk.SetDnsRequest{
		SubList: []*dynadotsdk.DnsSubRecord{
			{
				SubHost:      subHost,
				RecordType:   "TXT",
				RecordValue1: info.Value,
			},
		},
		TTL:                    lo.ToPtr(int64(d.config.TTL)),
		AddDnsToCurrentSetting: lo.ToPtr(true),
	}
	if _, err := d.client.SetDns(zone, request); err != nil {
		return fmt.Errorf("dynadot: error when create record: %w", err)
	}

	d.recordCacheMu.Lock()
	d.recordCache[token] = dnsRecordCacheEntry{Zone: zone, SubHost: subHost, Value: info.Value}
	d.recordCacheMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordCacheMu.Lock()
	record, ok := d.recordCache[token]
	d.recordCacheMu.Unlock()
	if !ok {
		return fmt.Errorf("dynadot: unknown record for '%s'", info.EffectiveFQDN)
	}

	// REF: https://www.dynadot.com/domain/api-document (remove_dns)
	request := &dynadotsdk.RemoveDnsRequest{
		SubList: []*dynadotsdk.DnsSubRecord{
			{
				SubHost:      record.SubHost,
				RecordType:   "TXT",
				RecordValue1: record.Value,
			},
		},
	}
	if _, err := d.client.RemoveDns(record.Zone, request); err != nil {
		return fmt.Errorf("dynadot: error when delete record: %w", err)
	}

	d.recordCacheMu.Lock()
	delete(d.recordCache, token)
	d.recordCacheMu.Unlock()

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
