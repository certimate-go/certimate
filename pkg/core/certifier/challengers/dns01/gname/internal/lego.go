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

	gnamesdk "github.com/certimate-go/certimate/pkg/sdk3rd/gname"
)

const (
	envNamespace = "GNAME_"

	EnvAppID  = envNamespace + "APP_ID"
	EnvAppKey = envNamespace + "APP_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AppID  string
	AppKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *gnamesdk.Client

	recordIDs   map[string]int64 // Key: ChallengeToken; Value: RecordID
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
	values, err := env.Get(EnvAppID, EnvAppKey)
	if err != nil {
		return nil, fmt.Errorf("gname: %w", err)
	}

	config := NewDefaultConfig()
	config.AppID = values[EnvAppID]
	config.AppKey = values[EnvAppKey]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gname: the configuration of the DNS provider is nil")
	}

	client, err := gnamesdk.NewClient(config.AppID, config.AppKey)
	if err != nil {
		return nil, fmt.Errorf("gname: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		recordIDs:   make(map[string]int64),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gname: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("gname: %w", err)
	}

	// REF: https://www.gname.vip/domain/api/dns/add
	request := &gnamesdk.AddDomainResolutionRequest{
		ZoneName:    lo.ToPtr(dns01.UnFqdn(authZone)),
		RecordType:  lo.ToPtr("TXT"),
		RecordName:  lo.ToPtr(subDomain),
		RecordValue: lo.ToPtr(info.Value),
		TTL:         lo.ToPtr(int32(d.config.TTL)),
	}
	response, err := d.client.AddDomainResolution(request)
	if err != nil {
		return fmt.Errorf("gname: error when create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token], _ = response.Data.Int64()
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("gname: could not find zone for domain %q: %w", domain, err)
	}

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("gname: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	// REF: https://www.gname.vip/domain/api/dns/del
	request := &gnamesdk.DeleteDomainResolutionRequest{
		ZoneName: lo.ToPtr(dns01.UnFqdn(authZone)),
		RecordID: lo.ToPtr(recordID),
	}
	_, err = d.client.DeleteDomainResolution(request)
	if err != nil {
		return fmt.Errorf("gname: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
