package internal

import (
	"errors"
	"fmt"
	"sync"
	"time"

	aliopen "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	aliesa "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	envNamespace = "ALICLOUDESA_"

	EnvAccessKey = envNamespace + "ACCESS_KEY"
	EnvSecretKey = envNamespace + "SECRET_KEY"
	EnvRegionID  = envNamespace + "REGION_ID"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	SecretID  string
	SecretKey string
	RegionID  string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int32
	HTTPTimeout        time.Duration
}

type DNSProvider struct {
	client *aliesa.Client
	config *Config

	siteIDs    map[string]int64
	siteIDsMtx sync.Mutex
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
	values, err := env.Get(EnvAccessKey, EnvSecretKey, EnvRegionID)
	if err != nil {
		return nil, fmt.Errorf("alicloud-esa: %w", err)
	}

	config := NewDefaultConfig()
	config.SecretID = values[EnvAccessKey]
	config.SecretKey = values[EnvSecretKey]
	config.RegionID = values[EnvRegionID]

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("alicloud-esa: the configuration of the DNS provider is nil")
	}

	client, err := aliesa.NewClient(&aliopen.Config{
		AccessKeyId:     tea.String(config.SecretID),
		AccessKeySecret: tea.String(config.SecretKey),
		Endpoint:        tea.String(fmt.Sprintf("esa.%s.aliyuncs.com", config.RegionID)),
	})
	if err != nil {
		return nil, fmt.Errorf("alicloud-esa: %w", err)
	}

	return &DNSProvider{
		client: client,
		config: config,

		siteIDs: make(map[string]int64),
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("alicloud-esa: could not find zone for domain %q: %w", domain, err)
	}

	siteName := dns01.UnFqdn(authZone)
	siteId, err := d.getSiteId(siteName)
	if err != nil {
		return fmt.Errorf("alicloud-esa: could not find site for zone %q: %w", siteName, err)
	}

	if err := d.addOrUpdateDNSRecord(siteId, dns01.UnFqdn(info.EffectiveFQDN), info.Value); err != nil {
		return fmt.Errorf("alicloud-esa: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("alicloud-esa: could not find zone for domain %q: %w", domain, err)
	}

	siteName := dns01.UnFqdn(authZone)
	siteId, err := d.getSiteId(siteName)
	if err != nil {
		return fmt.Errorf("alicloud-esa: could not find site for zone %q: %w", siteName, err)
	}

	if err := d.removeDNSRecord(siteId, dns01.UnFqdn(info.EffectiveFQDN), info.Value); err != nil {
		return fmt.Errorf("alicloud-esa: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getSiteId(siteName string) (int64, error) {
	d.siteIDsMtx.Lock()
	siteID, ok := d.siteIDs[siteName]
	d.siteIDsMtx.Unlock()
	if ok {
		return siteID, nil
	}

	pageNumber := 1
	pageSize := 500
	for {
		request := &aliesa.ListSitesRequest{
			SiteName:       tea.String(siteName),
			SiteSearchType: tea.String("exact"),
			PageNumber:     tea.Int32(int32(pageNumber)),
			PageSize:       tea.Int32(int32(pageSize)),
			AccessType:     tea.String("NS"),
		}
		response, err := d.client.ListSites(request)
		if err != nil {
			return 0, err
		}

		if response.Body == nil {
			break
		} else {
			for _, record := range response.Body.Sites {
				if tea.StringValue(record.SiteName) == siteName {
					d.siteIDsMtx.Lock()
					d.siteIDs[siteName] = *record.SiteId
					d.siteIDsMtx.Unlock()
					return *record.SiteId, nil
				}
			}

			if len(response.Body.Sites) < pageSize {
				break
			}

			pageNumber++
		}
	}

	return 0, errors.New("site not found")
}

func (d *DNSProvider) findDNSRecord(siteId int64, effectiveFQDN string, value string) (*aliesa.ListRecordsResponseBodyRecords, error) {
	pageNumber := 1
	pageSize := 500
	for {
		request := &aliesa.ListRecordsRequest{
			SiteId:          tea.Int64(siteId),
			Type:            tea.String("TXT"),
			RecordName:      tea.String(effectiveFQDN),
			RecordMatchType: tea.String("exact"),
			PageNumber:      tea.Int32(int32(pageNumber)),
			PageSize:        tea.Int32(int32(pageSize)),
		}
		response, err := d.client.ListRecords(request)
		if err != nil {
			return nil, err
		}

		if response.Body == nil {
			break
		} else {
			for _, record := range response.Body.Records {
				if tea.StringValue(record.RecordName) == effectiveFQDN && tea.StringValue(record.Data.Value) == value {
					return record, nil
				}
			}

			if len(response.Body.Records) < pageSize {
				break
			}

			pageNumber++
		}
	}

	return nil, nil
}

func (d *DNSProvider) addOrUpdateDNSRecord(siteId int64, effectiveFQDN, value string) error {
	record, err := d.findDNSRecord(siteId, effectiveFQDN, value)
	if err != nil {
		return err
	}

	if record != nil {
		return nil
	}

	request := &aliesa.CreateRecordRequest{
		SiteId:     tea.Int64(siteId),
		Type:       tea.String("TXT"),
		RecordName: tea.String(effectiveFQDN),
		Data: &aliesa.CreateRecordRequestData{
			Value: tea.String(value),
		},
		Ttl: tea.Int32(d.config.TTL),
	}
	_, err = d.client.CreateRecord(request)
	return err
}

func (d *DNSProvider) removeDNSRecord(siteId int64, effectiveFQDN, value string) error {
	record, err := d.findDNSRecord(siteId, effectiveFQDN, value)
	if err != nil {
		return err
	}

	if record == nil {
		return nil
	} else {
		request := &aliesa.DeleteRecordRequest{
			RecordId: record.RecordId,
		}
		_, err = d.client.DeleteRecord(request)
		return err
	}
}
