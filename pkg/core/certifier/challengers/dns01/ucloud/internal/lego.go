package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	ucloudsdk "github.com/certimate-go/certimate/pkg/sdk3rd/ucloud/udnr"
)

const (
	envNamespace = "UCLOUD_"

	EnvPublicKey  = envNamespace + "PUBLIC_KEY"
	EnvPrivateKey = envNamespace + "PRIVATE_KEY"
	EnvProjectID  = envNamespace + "PROJECT_ID"
	EnvBaseURL    = envNamespace + "BASE_URL"
	EnvRegion     = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	PublicKey  string
	PrivateKey string
	ProjectID  string
	BaseURL    string
	Region     string

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPTimeout        time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// 这里有意不使用 lego 提供的 ucloud 实现，
// 因为它无法配置客户端的访问地址。
type DNSProvider struct {
	config *Config
	client *ucloudsdk.UDNRClient
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvPublicKey, EnvPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("ucloud: %w", err)
	}

	config := NewDefaultConfig()
	config.PublicKey = values[EnvPublicKey]
	config.PrivateKey = values[EnvPrivateKey]
	config.ProjectID = env.GetOrFile(EnvProjectID)
	config.BaseURL = env.GetOrFile(EnvBaseURL)
	config.Region = env.GetOrFile(EnvRegion)

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ucloud: the configuration of the DNS provider is nil")
	}

	if config.PublicKey == "" || config.PrivateKey == "" {
		return nil, errors.New("ucloud: credentials missing")
	}

	credential := auth.NewCredential()
	credential.PublicKey = config.PublicKey
	credential.PrivateKey = config.PrivateKey

	cfg := ucloud.NewConfig()

	if config.ProjectID != "" {
		cfg.ProjectId = config.ProjectID
	}

	if config.BaseURL != "" {
		if strings.Contains(config.BaseURL, "://") {
			cfg.BaseUrl = config.BaseURL
		} else {
			cfg.BaseUrl = "https://" + config.BaseURL
		}
	}

	if config.Region != "" {
		cfg.Region = config.Region
	}

	return &DNSProvider{
		config: config,
		client: ucloudsdk.NewClient(&cfg, &credential),
	}, nil
}

func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ucloud: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docs.ucloud.cn/api/udnr-api/udnr_domain_dns_add
	addRequest := d.client.NewDomainDNSAddRequest()
	addRequest.Dn = ucloud.String(dns01.UnFqdn(authZone))
	addRequest.RecordName = ucloud.String(dns01.UnFqdn(info.EffectiveFQDN))
	addRequest.DnsType = ucloud.String("TXT")
	addRequest.Content = ucloud.String(info.Value)
	addRequest.TTL = ucloud.String(strconv.Itoa(d.config.TTL))
	addRequest.WithTimeout(d.config.HTTPTimeout)

	_, err = d.client.DomainDNSAdd(addRequest)
	if err != nil {
		return fmt.Errorf("ucloud: domain DNS add: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("ucloud: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docs.ucloud.cn/api/udnr-api/udnr_domain_dns_query
	queryRequest := d.client.NewDomainDNSQueryRequest()
	queryRequest.Dn = ucloud.String(dns01.UnFqdn(authZone))
	queryRequest.WithTimeout(d.config.HTTPTimeout)

	dom, err := d.client.DomainDNSQuery(queryRequest)
	if err != nil {
		return fmt.Errorf("ucloud: domain DNS query: %w", err)
	}

	for _, record := range dom.Data {
		if record.Type != "TXT" || record.Name != dns01.UnFqdn(info.EffectiveFQDN) || record.Content != info.Value {
			continue
		}

		// REF: https://docs.ucloud.cn/api/udnr-api/udnr_delete_dns_record
		deleteRequest := d.client.NewDNSRecordDeleteRequest()
		deleteRequest.Dn = ucloud.String(dns01.UnFqdn(authZone))
		deleteRequest.RecordName = ucloud.String(dns01.UnFqdn(info.EffectiveFQDN))
		deleteRequest.DnsType = ucloud.String(record.Type)
		deleteRequest.Content = ucloud.String(record.Content)
		deleteRequest.WithTimeout(d.config.HTTPTimeout)

		_, err = d.client.DNSRecordDelete(deleteRequest)
		if err != nil {
			return fmt.Errorf("ucloud: delete DNS record: %w", err)
		}
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
