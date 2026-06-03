package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	bpdns "github.com/byteplus-sdk/byteplus-sdk-golang/service/dns"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/samber/lo"
)

const (
	envNamespace = "BYTEPLUS_"

	EnvAccessKey = envNamespace + "ACCESSKEY"
	EnvSecretKey = envNamespace + "SECRETKEY"
	EnvRegion    = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultTTL = 600

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AccessKey string
	SecretKey string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 4*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 10*time.Second),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, time.Duration(bpdns.Timeout)*time.Second),
	}
}

type DNSProvider struct {
	client *bpdns.Client
	config *Config

	recordIDs   map[string]*string // Key: ChallengeToken; Value: RecordID
	recordIDsMu sync.Mutex
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("byteplus: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKey = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("byteplus: the configuration of the DNS provider is nil")
	}

	if config.AccessKey == "" || config.SecretKey == "" {
		return nil, errors.New("byteplus: missing credentials")
	}

	client := bpdns.InitDNSBytePlusClient()
	client.SetAccessKey(config.AccessKey)
	client.SetSecretKey(config.SecretKey)

	return &DNSProvider{
		config:    config,
		client:    client,
		recordIDs: make(map[string]*string),
	}, nil
}

func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	zoneInfo, err := d.findZone(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("byteplus: get zone ID: %w", err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, lo.FromPtr(zoneInfo.ZoneName))
	if err != nil {
		return fmt.Errorf("byteplus: %w", err)
	}

	record, err := d.client.CreateRecord(ctx, &bpdns.CreateRecordRequest{
		Host:  lo.ToPtr(subDomain),
		TTL:   lo.ToPtr(int64(d.config.TTL)),
		Type:  lo.ToPtr("TXT"),
		Value: lo.ToPtr(info.Value),
		ZID:   zoneInfo.ZID,
	})
	if err != nil {
		return fmt.Errorf("byteplus: error when create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = record.RecordID
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("byteplus: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	err := d.client.DeleteRecord(ctx, &bpdns.DeleteRecordRequest{
		RecordID: recordID,
	})
	if err != nil {
		return fmt.Errorf("byteplus: error when delete record: %w", err)
	}

	d.recordIDsMu.Lock()
	delete(d.recordIDs, token)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findZone(ctx context.Context, fqdn string) (bpdns.TopZoneResponse, error) {
	for domain := range dns01.UnFqdnDomainsSeq(fqdn) {
		lzr := &bpdns.ListZonesRequest{
			Key:        lo.ToPtr(domain),
			SearchMode: lo.ToPtr("exact"),
		}

		zones, err := d.client.ListZones(ctx, lzr)
		if err != nil {
			return bpdns.TopZoneResponse{}, fmt.Errorf("list zones: %w", err)
		}

		total := lo.FromPtr(zones.Total)

		if total == 0 || len(zones.Zones) == 0 {
			continue
		}

		if total > 1 {
			return bpdns.TopZoneResponse{}, fmt.Errorf("too many zone for %s", domain)
		}

		return zones.Zones[0], nil
	}

	return bpdns.TopZoneResponse{}, fmt.Errorf("zone no found for fqdn: %s", fqdn)
}
