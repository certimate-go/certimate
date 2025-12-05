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
	"gitlab.ecloud.com/ecloud/ecloudsdkclouddns"
	"gitlab.ecloud.com/ecloud/ecloudsdkclouddns/model"
	"gitlab.ecloud.com/ecloud/ecloudsdkcore/config"
)

const (
	envNamespace = "CMCCCLOUD_"

	EnvAccessKey = envNamespace + "ACCESS_KEY"
	EnvSecretKey = envNamespace + "SECRET_KEY"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvReadTimeout        = envNamespace + "READ_TIMEOUT"
	EnvConnectTimeout     = envNamespace + "CONNECT_TIMEOUT"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AccessKey string
	SecretKey string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	ReadTimeout        int
	ConnectTimeout     int
}

type DNSProvider struct {
	config *Config
	client *ecloudsdkclouddns.Client

	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

func NewDefaultConfig() *Config {
	return &Config{
		ReadTimeout:        env.GetOrDefaultInt(EnvReadTimeout, 30),
		ConnectTimeout:     env.GetOrDefaultInt(EnvConnectTimeout, 30),
		TTL:                env.GetOrDefaultInt(EnvTTL, 600),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 10*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKey, EnvSecretKey)
	if err != nil {
		return nil, fmt.Errorf("cmccecloud: %w", err)
	}

	cfg := NewDefaultConfig()
	cfg.AccessKey = values[EnvAccessKey]
	cfg.SecretKey = values[EnvSecretKey]

	return NewDNSProviderConfig(cfg)
}

func NewDNSProviderConfig(cfg *Config) (*DNSProvider, error) {
	if cfg == nil {
		return nil, errors.New("cmccecloud: the configuration of the DNS provider is nil")
	}

	client := ecloudsdkclouddns.NewClient(&config.Config{
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		// 资源池常量见: https://ecloud.10086.cn/op-help-center/doc/article/54462
		// 默认全局
		PoolId:         "CIDC-CORE-00",
		ReadTimeOut:    cfg.ReadTimeout,
		ConnectTimeout: cfg.ConnectTimeout,
	})

	return &DNSProvider{
		config:      cfg,
		client:      client,
		recordIDs:   make(map[string]string),
		recordIDsMu: sync.Mutex{},
	}, nil
}

func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zoneName, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("cmccecloud: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, zoneName)
	if err != nil {
		return fmt.Errorf("cmccecloud: %w", err)
	}

	cmccCreateRecordReq := &model.CreateRecordOpenapiRequest{
		CreateRecordOpenapiBody: &model.CreateRecordOpenapiBody{
			LineId:      "0", // 默认线路
			Rr:          subDomain,
			DomainName:  dns01.UnFqdn(zoneName),
			Description: "certimate acme",
			Type:        model.CreateRecordOpenapiBodyTypeEnumTxt,
			Value:       info.Value,
			Ttl:         lo.ToPtr(int32(d.config.TTL)),
		},
	}
	cmccCreateRecordResp, err := d.client.CreateRecordOpenapi(cmccCreateRecordReq)
	if err != nil {
		return fmt.Errorf("cmccecloud: error when create record: %w", err)
	} else if cmccCreateRecordResp.State != model.CreateRecordOpenapiResponseStateEnumOk {
		return fmt.Errorf("cmccecloud: failed to create record: unexpected response state: '%s', errcode: '%s', errmsg: '%s'", cmccCreateRecordResp.State, cmccCreateRecordResp.ErrorCode, cmccCreateRecordResp.ErrorMessage)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[token] = cmccCreateRecordResp.Body.RecordId
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[token]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("cmccecloud: unknown record ID for '%s'", info.EffectiveFQDN)
	}

	cmccDeleteRecordReq := &model.DeleteRecordOpenapiRequest{
		DeleteRecordOpenapiBody: &model.DeleteRecordOpenapiBody{
			RecordIdList: []string{recordID},
		},
	}
	cmccDeleteRecordResp, err := d.client.DeleteRecordOpenapi(cmccDeleteRecordReq)
	if err != nil {
		return fmt.Errorf("cmccecloud: error when delete record: %w", err)
	} else if cmccDeleteRecordResp.State != model.DeleteRecordOpenapiResponseStateEnumOk {
		return fmt.Errorf("cmccecloud: failed to delete record, unexpected response state: '%s', errcode: '%s', errmsg: '%s'", cmccDeleteRecordResp.State, cmccDeleteRecordResp.ErrorCode, cmccDeleteRecordResp.ErrorMessage)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
