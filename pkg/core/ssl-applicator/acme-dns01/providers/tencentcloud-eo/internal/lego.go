package internal

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

const (
	envNamespace = "TENCENTCLOUDEO_"

	EnvSecretID  = envNamespace + "SECRET_ID"
	EnvSecretKey = envNamespace + "SECRET_KEY"
	EnvZoneID    = envNamespace + "ZONE_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	SecretID  string
	SecretKey string
	ZoneID    string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int32
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	client *teo.Client
	config *Config
}

func NewDefaultConfig() *Config {
	return &Config{
		TTL:                int32(env.GetOrDefaultInt(EnvTTL, 300)),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvSecretID, EnvSecretKey, EnvZoneID)
	if err != nil {
		return nil, fmt.Errorf("tencentcloud-eo: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretID = values[EnvSecretID]
	config.SecretKey = values[EnvSecretKey]
	config.ZoneID = values[EnvSecretKey]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("tencentcloud-eo: the configuration of the DNS provider is nil")
	}

	credential := common.NewCredential(config.SecretID, config.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqTimeout = int(math.Round(config.HTTPTimeout.Seconds()))
	client, err := teo.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	if err := d.addDNSRecord(dns01.UnFqdn(info.EffectiveFQDN), info.Value); err != nil {
		return fmt.Errorf("tencentcloud-eo: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	if err := d.removeDNSRecord(dns01.UnFqdn(info.EffectiveFQDN), info.Value); err != nil {
		return fmt.Errorf("tencentcloud-eo: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) findDNSRecord(effectiveFQDN, value string) (*teo.DnsRecord, error) {
	pageOffset := 0
	pageLimit := 1000
	for {
		request := teo.NewDescribeDnsRecordsRequest()
		request.ZoneId = common.StringPtr(d.config.ZoneID)
		request.Offset = common.Int64Ptr(int64(pageOffset))
		request.Limit = common.Int64Ptr(int64(pageLimit))
		request.Filters = []*teo.AdvancedFilter{
			{
				Name:   common.StringPtr("type"),
				Values: []*string{common.StringPtr("TXT")},
			},
		}

		response, err := d.client.DescribeDnsRecords(request)
		if err != nil {
			return nil, err
		}

		if response.Response == nil {
			break
		} else {
			for _, record := range response.Response.DnsRecords {
				if *record.Name == effectiveFQDN && *record.Content == value {
					return record, nil
				}
			}

			if len(response.Response.DnsRecords) < pageLimit {
				break
			}

			pageOffset += len(response.Response.DnsRecords)
		}
	}

	return nil, nil
}

func (d *DNSProvider) addDNSRecord(effectiveFQDN, value string) error {
	request := teo.NewCreateDnsRecordRequest()
	request.ZoneId = common.StringPtr(d.config.ZoneID)
	request.Name = common.StringPtr(effectiveFQDN)
	request.Type = common.StringPtr("TXT")
	request.Content = common.StringPtr(value)
	request.TTL = common.Int64Ptr(int64(d.config.TTL))
	_, err := d.client.CreateDnsRecord(request)
	return err
}

func (d *DNSProvider) removeDNSRecord(effectiveFQDN, value string) error {
	record, err := d.findDNSRecord(effectiveFQDN, value)
	if err != nil {
		return err
	}

	if record == nil {
		return nil
	} else {
		request := teo.NewDeleteDnsRecordsRequest()
		request.ZoneId = common.StringPtr(d.config.ZoneID)
		request.RecordIds = []*string{record.RecordId}
		_, err = d.client.DeleteDnsRecords(request)
		return err
	}
}
