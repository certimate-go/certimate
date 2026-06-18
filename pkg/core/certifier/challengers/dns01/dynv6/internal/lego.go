package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
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

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *dynv6sdk.Client

	zoneIDs     map[string]int64 // Key: ZoneFQDN; Value: ZoneID
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
		return nil, fmt.Errorf("dynv6: the configuration of the DNS provider is nil")
	}

	client, err := dynv6sdk.NewClient(
		dynv6sdk.WithHttpToken(config.HTTPToken),
	)
	if err != nil {
		return nil, fmt.Errorf("dynv6: %w", err)
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

func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynv6: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dynv6: %w", err)
	}

	zoneInfo, err := d.findZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("dynv6: error when list zones: %w", err)
	}

	response, err := d.client.AddRecordWithContext(ctx, zoneInfo.ID, &dynv6sdk.AddRecordRequest{
		Type: lo.ToPtr("TXT"),
		Name: lo.ToPtr(subDomain),
		Data: lo.ToPtr(info.Value),
	})
	if err != nil {
		return fmt.Errorf("dynv6: error when create record: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[authZone] = zoneInfo.ID
	d.zoneIDsMu.Unlock()

	d.recordIDsMu.Lock()
	d.recordIDs[token] = response.ID
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dynv6: could not find zone for domain %q: %w", domain, err)
	}

	d.zoneIDsMu.Lock()
	zoneId, ok := d.zoneIDs[authZone]
	d.zoneIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("dynv6: unknown zone ID for '%s'", authZone)
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

	d.zoneIDsMu.Lock()
	delete(d.zoneIDs, authZone)
	d.zoneIDsMu.Unlock()

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, zoneName string) (*dynv6sdk.ZoneRecord, error) {
	zones, err := d.client.ListZonesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	for _, zone := range *zones {
		if dns01.UnFqdn(zone.Name) == dns01.UnFqdn(zoneName) {
			return zone, nil
		}
	}

	return nil, fmt.Errorf("could not find zone: '%s'", zoneName)
}
