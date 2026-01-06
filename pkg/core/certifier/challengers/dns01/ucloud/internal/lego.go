package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/udnr"
)

const (
	envNamespace = "UCLOUDUDNR_"

	EnvPublicKey  = envNamespace + "PUBLIC_KEY"
	EnvPrivateKey = envNamespace + "PRIVATE_KEY"
	EnvProjectId  = envNamespace + "PROJECT_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	PrivateKey string
	PublicKey  string
	ProjectId  string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	config *Config
	client *udnr.UDNRClient
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPrivateKey, EnvPublicKey, EnvProjectId)
	if err != nil {
		return nil, fmt.Errorf("ucloud: %w", err)
	}

	config := NewDefaultConfig()
	config.PrivateKey = values[EnvPrivateKey]
	config.PublicKey = values[EnvPublicKey]
	config.ProjectId = values[EnvProjectId]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ucloud: the configuration of the DNS provider is nil")
	}

	cfg := ucloud.NewConfig()
	cfg.Timeout = config.HTTPTimeout
	cfg.ProjectId = config.ProjectId
	credential := auth.NewCredential()
	credential.PrivateKey = config.PrivateKey
	credential.PublicKey = config.PublicKey
	client := udnr.NewClient(&cfg, &credential)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ucloud: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docs.ucloud.cn/api/udnr-api/udnr_domain_dns_add
	request := d.client.NewAddDomainDNSRequest()
	request.Dn = ucloud.String(dns01.UnFqdn(authZone))
	request.DnsType = ucloud.String("TXT")
	request.RecordName = ucloud.String(dns01.UnFqdn(info.EffectiveFQDN))
	request.Content = ucloud.String(info.Value)
	request.TTL = ucloud.String(fmt.Sprintf("%d", d.config.TTL))
	if _, err := d.client.AddDomainDNS(request); err != nil {
		return fmt.Errorf("ucloud: error when create record: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ucloud: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docs.ucloud.cn/api/udnr-api/udnr_domain_dns_query
	request := d.client.NewQueryDomainDNSRequest()
	request.Dn = ucloud.String(dns01.UnFqdn(authZone))
	response, err := d.client.QueryDomainDNS(request)
	if err != nil {
		return fmt.Errorf("ucloud: error when list records: %w", err)
	}

	// REF: https://docs.ucloud.cn/api/udnr-api/udnr_delete_dns_record
	for _, record := range response.Data {
		if record.DnsType == "TXT" && record.RecordName == dns01.UnFqdn(info.EffectiveFQDN) && record.Content == info.Value {
			delreq := d.client.NewDeleteDomainDNSRequest()
			delreq.Dn = ucloud.String(dns01.UnFqdn(authZone))
			delreq.DnsType = ucloud.String(record.DnsType)
			delreq.RecordName = ucloud.String(record.RecordName)
			delreq.Content = ucloud.String(record.Content)
			_, err := d.client.DeleteDomainDNS(delreq)
			if err != nil {
				return fmt.Errorf("ucloud: error when delete record: %w", err)
			}
			break
		}
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
