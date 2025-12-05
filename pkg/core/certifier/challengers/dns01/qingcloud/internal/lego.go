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

	qingcloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/qingcloud/dns"
)

const (
	envNamespace = "DNSLA_"

	EnvAccessKey    = envNamespace + "ACCESS_KEY"
	EnvAccessSecret = envNamespace + "ACCESS_SECRET"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AccessKey    string
	AccessSecret string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *qingcloudsdk.Client

	recordIDs   map[string]*int64
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
	values, err := env.Get(EnvAccessKey, EnvAccessSecret)
	if err != nil {
		return nil, fmt.Errorf("qingcloud: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKey = values[EnvAccessKey]
	config.AccessSecret = values[EnvAccessSecret]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("qingcloud: the configuration of the DNS provider is nil")
	}

	client, err := qingcloudsdk.NewClient(config.AccessKey, config.AccessSecret)
	if err != nil {
		return nil, fmt.Errorf("qingcloud: %w", err)
	} else {
		client.SetTimeout(config.HTTPTimeout)
	}

	return &DNSProvider{
		config:      config,
		client:      client,
		recordIDs:   make(map[string]*int64),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("qingcloud: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docsv4.qingcloud.com/user_guide/development_docs/api/api_list/dns/record/#_createrecord
	qingcloudCreateRecordReq := &qingcloudsdk.CreateRecordRequest{
		ZoneName:   lo.ToPtr(authZone),
		DomainName: lo.ToPtr(info.EffectiveFQDN),
		ViewId:     lo.ToPtr(int32(0)),
		Type:       lo.ToPtr("TXT"),
		Ttl:        lo.ToPtr(int32(d.config.TTL)),
		Records: []*qingcloudsdk.CreateRecordRequestRecord{
			{
				Values: []*qingcloudsdk.CreateRecordRequestRecordValue{
					{
						Value:  lo.ToPtr(info.Value),
						Status: lo.ToPtr(int32(1)),
					},
				},
				Weight: lo.ToPtr(int32(0)),
			},
		},
		Mode:      lo.ToPtr(int32(1)),
		AutoMerge: lo.ToPtr(int32(1)),
	}
	qingcloudCreateRecordResp, err := d.client.CreateRecord(qingcloudCreateRecordReq)
	if err != nil {
		return fmt.Errorf("qingcloud: error when create record: %w", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = qingcloudCreateRecordResp.DomainRecordId
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("qingcloud: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	// REF: https://docsv4.qingcloud.com/user_guide/development_docs/api/api_list/dns/record/#_deleterecord
	if _, err := d.client.DeleteRecord([]*int64{recordID}); err != nil {
		return fmt.Errorf("qingcloud: error when delete record: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
