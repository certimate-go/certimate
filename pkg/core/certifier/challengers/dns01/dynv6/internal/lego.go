package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/libdns/dynv6"
	"github.com/libdns/libdns"
)

const (
	envNamespace = "DYNV6_"

	EnvHTTPToken = envNamespace + "HTTP_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	HTTPToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

type DNSProvider struct {
	config *Config
	client *dynv6.Provider
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvHTTPToken)
	if err != nil {
		return nil, fmt.Errorf("dynv6: %w", err)
	}

	config := NewDefaultConfig()
	config.HTTPToken = values[EnvHTTPToken]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dynv6: the configuration of the DNS provider is nil")
	}

	client := &dynv6.Provider{Token: config.HTTPToken}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynv6: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dynv6: %w", err)
	}

	if _, err := d.client.AppendRecords(context.Background(), dns01.UnFqdn(authZone), []libdns.Record{&libdns.TXT{
		Name: subDomain,
		Text: info.Value,
		TTL:  time.Duration(d.config.TTL),
	}}); err != nil {
		return fmt.Errorf("dynv6: error when create record: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynv6: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dynv6: %w", err)
	}

	record, err := d.findDNSRecord(dns01.UnFqdn(authZone), subDomain, info.Value)
	if err != nil {
		return fmt.Errorf("dynv6: error when find record: %w", err)
	}

	if _, err := d.client.DeleteRecords(context.Background(), dns01.UnFqdn(authZone), []libdns.Record{record}); err != nil {
		return fmt.Errorf("dynv6: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDNSRecord(zoneName, subDomain, tokenValue string) (libdns.Record, error) {
	records, err := d.client.GetRecords(context.Background(), zoneName)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		rr := record.RR()
		if rr.Type == "TXT" && rr.Name == subDomain && rr.Data == tokenValue {
			return record, nil
		}
	}

	return nil, errors.New("could not find record")
}
