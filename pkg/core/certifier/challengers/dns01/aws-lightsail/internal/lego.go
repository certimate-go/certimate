package internal

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

const (
	envNamespace = "LIGHTSAIL_"

	EnvRegion = envNamespace + "REGION"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const maxRetries = 5

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

type DNSProvider struct {
	client *lightsail.Client
	config *Config
}

func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Region = env.GetOrDefaultString(EnvRegion, "us-east-1")

	return NewDNSProviderConfig(config)
}

func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("lightsail: the configuration of the DNS provider is nil")
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, config.SessionToken)),
		awsconfig.WithRegion(config.Region),
		awsconfig.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(options *retry.StandardOptions) {
				options.MaxAttempts = maxRetries
				options.Backoff = retry.BackoffDelayerFunc(func(attempt int, err error) (time.Duration, error) {
					retryCount := min(attempt, 7)
					delay := (1 << uint(retryCount)) * (rand.IntN(50) + 200)
					return time.Duration(delay) * time.Millisecond, nil
				})
			})
		}),
	)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		config: config,
		client: lightsail.NewFromConfig(cfg),
	}, nil
}

func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("lightsail: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	if _, err := d.client.CreateDomainEntry(context.Background(), &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(dns01.UnFqdn(authZone)),
		DomainEntry: &awstypes.DomainEntry{
			Type:   aws.String("TXT"),
			Name:   aws.String(subDomain),
			Target: aws.String(strconv.Quote(info.Value)),
		},
	}); err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	return nil
}

func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("lightsail: could not find zone for domain %q: %w", domain, err)
	}

	subDomain, err := dns01.ExtractSubDomain(info.EffectiveFQDN, authZone)
	if err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	if _, err := d.client.DeleteDomainEntry(context.Background(), &lightsail.DeleteDomainEntryInput{
		DomainName: aws.String(dns01.UnFqdn(authZone)),
		DomainEntry: &awstypes.DomainEntry{
			Type:   aws.String("TXT"),
			Name:   aws.String(subDomain),
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
