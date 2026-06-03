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

	conohavpssdk "github.com/certimate-go/certimate/pkg/sdk3rd/conoha/vps/v3"
)

const (
	envNamespace = "CONOHAV3_"

	EnvAPIUserID   = envNamespace + "API_USER_ID"
	EnvAPIUserName = envNamespace + "API_USER_NAME"
	EnvAPIPassword = envNamespace + "API_PASSWORD"
	EnvTenantID    = envNamespace + "TENANT_ID"
	EnvTenantName  = envNamespace + "TENANT_NAME"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	UserID     string
	UserName   string
	Password   string
	TenantID   string
	TenantName string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *conohavpssdk.Client

	zoneIDs     map[string]string // Key: ZoneFQDN; Value: ZoneUUID
	zoneIDsMu   sync.Mutex
	recordIDs   map[string]string // Key: ChallengeToken; Value: RecordUUID
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
	values, err := env.Get(EnvAPIUserID, EnvAPIUserName, EnvAPIPassword, EnvTenantID, EnvTenantName)
	if err != nil {
		return nil, fmt.Errorf("conohavpsv3: %w", err)
	}

	config := NewDefaultConfig()
	config.UserID = values[EnvAPIUserID]
	config.UserName = values[EnvAPIUserName]
	config.Password = values[EnvAPIPassword]
	config.TenantID = values[EnvTenantID]
	config.TenantName = values[EnvTenantName]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("conohavpsv3: the configuration of the DNS provider is nil")
	}

	client, err := conohavpssdk.NewClient(config.UserID, config.UserName, config.Password, config.TenantID, config.TenantName)
	if err != nil {
		return nil, fmt.Errorf("conohavpsv3: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		zoneIDs:     make(map[string]string),
		zoneIDsMu:   sync.Mutex{},
		recordIDs:   make(map[string]string),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("conohavpsv3: could not find zone for domain %q: %w", domain, err)
	}

	zoneInfo, err := d.findZone(ctx, authZone)
	if err != nil {
		return fmt.Errorf("conohavpsv3: error when list zones: %w", err)
	}

	// REF: https://doc.conoha.jp/reference/api-vps3/api-dns-vps3/dnsaas-create_record-v3/
	response, err := d.client.DnsCreateRecordWithContext(ctx, zoneInfo.UUID, &conohavpssdk.DnsCreateRecordRequest{
		Name: lo.ToPtr(info.EffectiveFQDN),
		Type: lo.ToPtr("TXT"),
		Data: lo.ToPtr(info.Value),
		TTL:  lo.ToPtr(d.config.TTL),
	})
	if err != nil {
		return fmt.Errorf("conohavpsv3: error when create record: %w", err)
	}

	d.zoneIDsMu.Lock()
	d.zoneIDs[authZone] = zoneInfo.UUID
	d.zoneIDsMu.Unlock()

	d.recordIDsMu.Lock()
	d.recordIDs[token] = response.UUID
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("conohavpsv3: could not find zone for domain %q: %w", domain, err)
	}

	d.zoneIDsMu.Lock()
	zoneId, ok := d.zoneIDs[authZone]
	d.zoneIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("conohavpsv3: unknown zone ID for '%s'", authZone)
	}

	d.recordIDsMu.Lock()
	recordId, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("conohavpsv3: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	if _, err := d.client.DnsDeleteRecordWithContext(ctx, zoneId, recordId); err != nil {
		return fmt.Errorf("conohavpsv3: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, zoneName string) (*conohavpssdk.DnsDomainRecord, error) {
	offset := 0
	limit := 10
	for {
		// REF: https://doc.conoha.jp/reference/api-vps3/api-dns-vps3/dnsaas-get_domains_list-v3/
		request := &conohavpssdk.DnsGetDomainsListRequest{
			Offset: lo.ToPtr(offset),
			Limit:  lo.ToPtr(limit),
		}
		response, err := d.client.DnsGetDomainsListWithContext(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, domainItem := range response.Domains {
			if dns01.UnFqdn(domainItem.Name) == dns01.UnFqdn(zoneName) {
				return domainItem, nil
			}
		}

		if len(response.Domains) < limit || offset+limit >= response.TotalCount {
			break
		}

		offset += limit
	}

	return nil, fmt.Errorf("could not find zone '%s'", zoneName)
}
