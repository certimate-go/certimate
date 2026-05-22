package internal

import (
	"fmt"
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
}

type dnsRecordCacheEntry struct {
	Zone    string
	SubHost string
	Value   string
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
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
		config: config,
		client: client,
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynadot: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dynadot: %w", err)
	}

	// REF: https://www.dynadot.com/domain/api-document (set_dns)
	request := &dynadotsdk.SetDnsRequest{
		SubList: []*dynadotsdk.DnsSubRecord{
			{
				SubHost:      subDomain,
				RecordType:   "TXT",
				RecordValue1: info.Value,
			},
		},
		TTL:                    lo.ToPtr(int64(d.config.TTL)),
		AddDnsToCurrentSetting: lo.ToPtr(true),
	}
	if _, err := d.client.SetDns(dns01.UnFqdn(authZone), request); err != nil {
		return fmt.Errorf("dynadot: error when create record: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynadot: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dynadot: %w", err)
	}

	// REF: https://www.dynadot.com/domain/api-document (remove_dns)
	request := &dynadotsdk.RemoveDnsRequest{
		SubList: []*dynadotsdk.DnsSubRecord{
			{
				SubHost:      subDomain,
				RecordType:   "TXT",
				RecordValue1: info.Value,
			},
		},
	}
	if _, err := d.client.RemoveDns(dns01.UnFqdn(authZone), request); err != nil {
		return fmt.Errorf("dynadot: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
