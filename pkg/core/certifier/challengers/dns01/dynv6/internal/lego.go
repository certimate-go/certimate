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

	dynv6sdk "github.com/certimate-go/certimate/pkg/sdk3rd/dynv6"
)

const (
	envNamespace = "DYNV6_"

	EnvHTTPToken = envNamespace + "HTTP_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	HTTPToken string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *dynv6sdk.Client

	zoneIDs     map[string]int64 // Key: ZoneName; Value: ZoneID
	zoneIDsMu   sync.Mutex
	recordIDs   map[string]int64 // Key: ChallengeToken; Value: RecordID
	recordIDsMu sync.Mutex
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

	client, err := dynv6sdk.NewClient(config.HTTPToken)
	if err != nil {
		return nil, fmt.Errorf("dnsexit: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		zoneIDs:     make(map[string]int64),
		zoneIDsMu:   sync.Mutex{},
		recordIDs:   make(map[string]int64),
		recordIDsMu: sync.Mutex{},
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

	zone, err := d.findZone(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("dynv6: error when list zones: %w", err)
	}

	// REF: https://dynv6.github.io/api-spec/#tag/records/operation/addRecord
	response, err := d.client.AddRecord(zone.ID, &dynv6sdk.AddRecordRequest{
		Type: lo.ToPtr("TXT"),
		Name: lo.ToPtr(subDomain),
		Data: lo.ToPtr(info.Value),
	})
	if err != nil {
		return fmt.Errorf("dynv6: error when create record: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[zone.Name] = zone.ID
	d.zoneIDsMu.Unlock()

	d.recordIDsMu.Lock()
	d.recordIDs[token] = response.ID
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynv6: could not find zone for domain %q: %w", domain, err)
	}

	d.zoneIDsMu.Lock()
	zoneId, ok := d.zoneIDs[dns01.UnFqdn(authZone)]
	d.zoneIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("dynv6: unknown zone ID for '%s'", dns01.UnFqdn(authZone))
	}

	d.recordIDsMu.Lock()
	recordId, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("dynv6: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	if _, err := d.client.DeleteRecord(zoneId, recordId); err != nil {
		return fmt.Errorf("dynv6: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(zoneName string) (*dynv6sdk.ZoneRecord, error) {
	// REF: https://dynv6.github.io/api-spec/#tag/zones/operation/findZones
	zones, err := d.client.ListZones()
	if err != nil {
		return nil, err
	}

	for _, zone := range *zones {
		if zone.Name == zoneName {
			return zone, nil
		}
	}

	return nil, fmt.Errorf("could not find zone: '%s'", zoneName)
}
