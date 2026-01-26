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

	dnscomsdk "github.com/certimate-go/certimate/pkg/sdk3rd/51dnscom"
)

const (
	envNamespace = "51DNSCOM_"
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
	client *dnscomsdk.Client

	recordCache   map[string]dnsRecordCacheEntry // Key: ChallengeToken
	recordCacheMu sync.Mutex
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
		return nil, fmt.Errorf("51dnscom: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("51dnscom: the configuration of the DNS provider is nil")
	}

	client, err := dnscomsdk.NewClient(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("51dnscom: %w", err)
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
		return fmt.Errorf("51dnscom: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("51dnscom: %w", err)
	}

	zone, err := d.findZone(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("51dnscom: error when list zones: %w", err)
	}

	// REF: https://www.51dns.com/document/api/4/12.html
	request := &dnscomsdk.RecordCreateRequest{
		DomainID: lo.ToPtr(zone.DomainID.String()),
		Type:     lo.ToPtr("TXT"),
		Host:     lo.ToPtr(subDomain),
		Value:    lo.ToPtr(info.Value),
		TTL:      lo.ToPtr(int32(d.config.TTL)),
	}
	response, err := d.client.RecordCreate(request)
	if err != nil {
		return fmt.Errorf("51dnscom: error when create record: %w", err)
	}

	d.recordCacheMu.Lock()
	d.recordCache[token] = dnsRecordCacheEntry{DomainID: zone.DomainID.String(), RecordID: response.Data.RecordID.String()}
	d.recordCacheMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordCacheMu.Lock()
	record, ok := d.recordCache[token]
	d.recordCacheMu.Unlock()
	if !ok {
		return fmt.Errorf("51dnscom: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	// REF: https://www.51dns.com/document/api/4/27.html
	request := &dnscomsdk.RecordRemoveRequest{
		DomainID: lo.ToPtr(record.DomainID),
		RecordID: lo.ToPtr(record.RecordID),
	}
	if _, err := d.client.RecordRemove(request); err != nil {
		return fmt.Errorf("51dnscom: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

type dnsRecordCacheEntry struct {
	DomainID string
	RecordID string
}

func (d *DNSProvider) findZone(zoneName string) (*dnscomsdk.DomainRecord, error) {
	page := 1
	pageSize := 10
	for {
		// REF: https://www.51dns.com/document/api/74/88.html
		request := &dnscomsdk.DomainListRequest{
			Page:     lo.ToPtr(int32(page)),
			PageSize: lo.ToPtr(int32(pageSize)),
		}
		response, err := d.client.DomainList(request)
		if err != nil {
			return nil, err
		}

		if response.Data == nil {
			break
		}

		for _, domainItem := range response.Data.Data {
			if domainItem.Domain == zoneName {
				return domainItem, nil
			}
		}

		if len(response.Data.Data) < pageSize || response.Data.PageCount <= int32(page) {
			break
		}

		page++
	}

	return nil, fmt.Errorf("could not find zone '%s'", zoneName)
}
