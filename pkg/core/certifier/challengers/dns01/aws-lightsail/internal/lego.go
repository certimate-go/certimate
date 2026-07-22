package internal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"

	awslightsailsdk "github.com/certimate-go/certimate/pkg/sdk3rd/aws/lightsail"
)

const (
	envNamespace = "LIGHTSAIL_"

	EnvRegion = envNamespace + "REGION"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// 这里有意不使用 lego 提供的 lightsail 实现，
// 因为它只支持单个域，无法签发多域名证书。
type DNSProvider struct {
	client *awslightsailsdk.Client
	config *Config
}

func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Region = env.GetOrDefaultString(EnvRegion, "us-east-1")

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("lightsail: the configuration of the DNS provider is nil")
	}

	client, err := awslightsailsdk.NewClient(
		awslightsailsdk.WithAkSk(config.AccessKeyID, config.SecretAccessKey),
		awslightsailsdk.WithRegion(config.Region),
	)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

func (d *DNSProvider) Present(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("lightsail: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_CreateDomainEntry.html
	if _, err := d.client.CreateDomainEntryWithContext(ctx, &awslightsailsdk.CreateDomainEntryRequest{
		DomainName: aws.String(dns01.UnFqdn(authZone)),
		DomainEntry: &types.DomainEntry{
			Type:   aws.String("TXT"),
			Name:   aws.String(info.EffectiveFQDN),
			Target: aws.String(strconv.Quote(info.Value)),
		},
	}); err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("lightsail: could not find zone for domain %q: %w", domain, err)
	}

	// REF: https://docs.aws.amazon.com/lightsail/2016-11-28/api-reference/API_DeleteDomainEntry.html
	if _, err := d.client.DeleteDomainEntryWithContext(ctx, &awslightsailsdk.DeleteDomainEntryRequest{
		DomainName: aws.String(dns01.UnFqdn(authZone)),
		DomainEntry: &types.DomainEntry{
			Type:   aws.String("TXT"),
			Name:   aws.String(info.EffectiveFQDN),
			Target: aws.String(strconv.Quote(info.Value)),
		},
	}); err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	return nil
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
