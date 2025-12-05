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

	ctyundns "github.com/certimate-go/certimate/pkg/sdk3rd/ctyun/dns"
)

const (
	envNamespace = "CTYUNSMARTDNS_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AccessKeyId     string
	SecretAccessKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	client *ctyundns.Client
	config *Config

	recordIDs   map[string]int32
	recordIDsMu sync.Mutex
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKeyID, EnvSecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("ctyun: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKeyId = values[EnvAccessKeyID]
	config.SecretAccessKey = values[EnvSecretAccessKey]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ctyun: the configuration of the DNS provider is nil")
	}

	client, err := ctyundns.NewClient(config.AccessKeyId, config.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("ctyun: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		client:      client,
		config:      config,
		recordIDs:   make(map[string]int32),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ctyun: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("ctyun: %w", err)
	}

	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=122&api=11259&data=181&isNormal=1&vid=259
	ctyunAddRecordReq := &ctyundns.AddRecordRequest{
		Domain:   lo.ToPtr(dns01.UnFqdn(authZone)),
		Host:     lo.ToPtr(subDomain),
		Type:     lo.ToPtr("TXT"),
		LineCode: lo.ToPtr("Default"),
		Value:    lo.ToPtr(info.Value),
		State:    lo.ToPtr(int32(1)),
		TTL:      lo.ToPtr(int32(d.config.TTL)),
	}
	ctyunAddRecordResp, err := d.client.AddRecord(ctyunAddRecordReq)
	if err != nil {
		return fmt.Errorf("ctyun: error when create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = ctyunAddRecordResp.ReturnObj.RecordId
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("tencentcloud-eo: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	// REF: https://eop.ctyun.cn/ebp/ctapiDocument/search?sid=122&api=11262&data=181&isNormal=1&vid=259
	ctyunDeleteRecordReq := &ctyundns.DeleteRecordRequest{
		RecordId: lo.ToPtr(recordID),
	}
	if _, err := d.client.DeleteRecord(ctyunDeleteRecordReq); err != nil {
		return fmt.Errorf("ctyun: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
