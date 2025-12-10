package internal

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/samber/lo"

	dnslasdk "github.com/certimate-go/certimate/pkg/sdk3rd/dnsla"
)

const (
	envNamespace = "DNSLA_"

	EnvAPIId     = envNamespace + "API_ID"
	EnvAPISecret = envNamespace + "API_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	APIId     string
	APISecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *dnslasdk.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 5*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIId, EnvAPISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsla: %w", err)
	}

	config := NewDefaultConfig()
	config.APIId = values[EnvAPIId]
	config.APISecret = values[EnvAPISecret]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsla: the configuration of the DNS provider is nil")
	}

	client, err := dnslasdk.NewClient(config.APIId, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsla: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		recordIDs:   make(map[string]string),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dnsla: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("dnsla: %w", err)
	}

	zone, err := d.findZone(dns01.UnFqdn(authZone))
	if err != nil {
		return fmt.Errorf("dnsla: error when list zones: %w", err)
	}

	// REF: https://www.dnsla.cn/docs/ApiDoc
	request := &dnslasdk.CreateRecordRequest{
		DomainId: lo.ToPtr(zone.Id),
		Type:     lo.ToPtr(int32(16)),
		Host:     lo.ToPtr(subDomain),
		Data:     lo.ToPtr(info.Value),
		Ttl:      lo.ToPtr(int32(d.config.TTL)),
	}
	response, err := d.client.CreateRecord(request)
	if err != nil {
		return fmt.Errorf("dnsla: error when create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = response.Data.Id
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("dnsla: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	// REF: https://www.dnsla.cn/docs/ApiDoc
	if _, err := d.client.DeleteRecord(recordID); err != nil {
		return fmt.Errorf("dnsla: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(zoneName string) (*dnslasdk.DomainRecord, error) {
	pageIndex := 1
	pageSize := 100
	for {
		// REF: https://www.dnsla.cn/docs/ApiDoc
		request := &dnslasdk.ListDomainsRequest{
			PageIndex: lo.ToPtr(int32(pageIndex)),
			PageSize:  lo.ToPtr(int32(pageSize)),
		}
		response, err := d.client.ListDomains(request)
		if err != nil {
			return nil, err
		}

		if response.Data == nil {
			break
		}

		for _, domainItem := range response.Data.Results {
			if strings.TrimRight(domainItem.Domain, ".") == zoneName || strings.TrimRight(domainItem.DisplayDomain, ".") == zoneName {
				return domainItem, nil
			}
		}

		if len(response.Data.Results) < pageSize {
			break
		}

		pageIndex++
	}

	return nil, fmt.Errorf("could not find zone '%s'", zoneName)
}
