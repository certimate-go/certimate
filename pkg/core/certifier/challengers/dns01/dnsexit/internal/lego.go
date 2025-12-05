package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/samber/lo"

	dnsexitsdk "github.com/certimate-go/certimate/pkg/sdk3rd/dnsexit"
)

const (
	envNamespace = "DNSEXIT_"

	EnvAPIKey = envNamespace + "APIKEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	APIKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 0),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

type DNSProvider struct {
	config *Config
	client *dnsexitsdk.Client
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("dnsexit: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsexit: the configuration of the DNS provider is nil")
	}

	client, err := dnsexitsdk.NewClient(config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("dnsexit: %w", err)
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
		return fmt.Errorf("dnsexit: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dnsexit: %w", err)
	}

	// REF: https://dnsexit.com/dns/dns-api/#example-add-new
	request := &dnsexitsdk.DnsRecordRequest{
		Domain: lo.ToPtr(dns01.UnFqdn(authZone)),
		Add: &dnsexitsdk.DnsRecord{
			Type:      lo.ToPtr("TXT"),
			Name:      lo.ToPtr(subDomain),
			Content:   lo.ToPtr(info.Value),
			TTL:       lo.ToPtr(d.config.TTL),
			Overwrite: lo.ToPtr(true),
		},
	}
	if _, err := d.client.DnsRecord(request); err != nil {
		return fmt.Errorf("dnsexit: error when create record: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnsexit: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dnsexit: %w", err)
	}

	// REF: https://dnsexit.com/dns/dns-api/#delete-a-record
	request := &dnsexitsdk.DnsRecordRequest{
		Domain: lo.ToPtr(dns01.UnFqdn(authZone)),
		Delete: &dnsexitsdk.DnsRecord{
			Type: lo.ToPtr("TXT"),
			Name: lo.ToPtr(subDomain),
		},
	}
	if _, err := d.client.DnsRecord(request); err != nil {
		return fmt.Errorf("dnsexit: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
