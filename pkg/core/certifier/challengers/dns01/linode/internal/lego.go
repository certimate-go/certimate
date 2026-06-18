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

	linodesdk "github.com/certimate-go/certimate/pkg/sdk3rd/linode"
)

// Environment variables names.
const (
	envNamespace = "LINODE_"

	EnvAccessToken = envNamespace + "ACCESS_TOKEN"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	minTTL             = 300
	dnsUpdateFreqMins  = 15
	dnsUpdateFudgeSecs = 120
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AccessToken string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, minTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 15*time.Second),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// 这里有意不使用 lego 提供的 linode 实现，
// 因为它引入了 Linode SDK，会导致依赖包体积增大。
type DNSProvider struct {
	config *Config
	client *linodesdk.Client

	domainIDs   map[string]int // Key: ZoneFQDN; Value: DomainID
	domainIDsMu sync.Mutex
	recordIDs   map[string]int // Key: ChallengeToken; Value: RecordID
	recordIDsMu sync.Mutex
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessToken)
	if err != nil {
		return nil, fmt.Errorf("linode: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessToken = values[EnvAccessToken]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("linode: the configuration of the DNS provider is nil")
	}

	client, err := linodesdk.NewClient(
		linodesdk.WithAccessToken(config.AccessToken),
	)
	if err != nil {
		return nil, fmt.Errorf("linode: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		domainIDs:   make(map[string]int),
		domainIDsMu: sync.Mutex{},
		recordIDs:   make(map[string]int),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("linode: could not find zone for domain %q: %w", domain, err)
	}

	zoneInfo, err := d.findZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("linode: error when list domains: %w", err)
	}

	response, err := d.client.CreateDomainRecordWithContext(ctx, lo.FromPtr(zoneInfo.ID), &linodesdk.CreateDomainRecordRequest{
		Name:   lo.ToPtr(dns01.UnFqdn(info.EffectiveFQDN)),
		Type:   lo.ToPtr("TXT"),
		Target: lo.ToPtr(info.Value),
		TTLSec: lo.ToPtr(d.config.TTL),
	})
	if err != nil {
		return fmt.Errorf("linode: error when create record: %w", err)
	}

	d.domainIDsMu.Lock()
	d.domainIDs[authZone] = lo.FromPtr(zoneInfo.ID)
	d.domainIDsMu.Unlock()

	d.recordIDsMu.Lock()
	d.recordIDs[token] = lo.FromPtr(response.ID)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("linode: could not find zone for domain %q: %w", domain, err)
	}

	d.domainIDsMu.Lock()
	domainId, ok := d.domainIDs[authZone]
	d.domainIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("linode: unknown domain ID for '%s'", authZone)
	}

	d.recordIDsMu.Lock()
	recordId, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("linode: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	if _, err := d.client.DeleteDomainRecordWithContext(ctx, domainId, recordId); err != nil {
		return fmt.Errorf("linode: error when delete record: %w", err)
	}

	d.domainIDsMu.Lock()
	delete(d.domainIDs, authZone)
	d.domainIDsMu.Unlock()

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) Timeout() (time.Duration, time.Duration) {
	timeout := d.config.PropagationTimeout
	if d.config.PropagationTimeout <= 0 {
		// Since Linode only updates their zone files every X minutes, we need
		// to figure out how many minutes we have to wait until we hit the next
		// interval of X.  We then wait another couple of minutes, just to be
		// safe.  Hopefully at some point during all of this, the record will
		// have propagated throughout Linode's network.
		minsRemaining := dnsUpdateFreqMins - (time.Now().Minute() % dnsUpdateFreqMins)

		timeout = (time.Duration(minsRemaining) * time.Minute) +
			(minTTL * time.Second) +
			(dnsUpdateFudgeSecs * time.Second)
	}

	return timeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, zoneName string) (*linodesdk.Domain, error) {
	page := 1
	pageSize := 100
	for {
		request := &linodesdk.ListDomainsRequest{
			Page:     lo.ToPtr(page),
			PageSize: lo.ToPtr(pageSize),
		}
		response, err := d.client.ListDomainsWithContext(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, domainItem := range response.Data {
			if dns01.UnFqdn(lo.FromPtr(domainItem.Domain)) == dns01.UnFqdn(zoneName) {
				return domainItem, nil
			}
		}

		if len(response.Data) < pageSize || (page*pageSize) >= response.Results {
			break
		}

		page++
	}

	return nil, fmt.Errorf("could not find zone '%s'", zoneName)
}
