package internal

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/log"

	"github.com/certimate-go/certimate/internal/tools/s3"
)

var _ challenge.Provider = (*HTTPProvider)(nil)

type Config struct {
	s3.Config

	Bucket string
}

func NewDefaultConfig() *Config {
	defaultCfg := s3.NewDefaultConfig()

	return &Config{
		Config: *defaultCfg,
	}
}

type HTTPProvider struct {
	config *Config
}

func NewHTTPProviderConfig(config *Config) (*HTTPProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	return &HTTPProvider{
		config: config,
	}, nil
}

func (p *HTTPProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createS3Client()
	if err != nil {
		return fmt.Errorf("s3: failed to create S3 client: %w", err)
	}

	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	if err := client.PutObjectString(ctx, p.config.Bucket, objectKey, keyAuth); err != nil {
		return fmt.Errorf("s3: failed to upload file for HTTP challenge: %w", err)
	}

	log.Info("s3: authz file uploaded", slog.String("bucket", p.config.Bucket), slog.String("object", objectKey))

	return nil
}

func (p *HTTPProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	client, err := p.createS3Client()
	if err != nil {
		return fmt.Errorf("s3: failed to create S3 client: %w", err)
	}

	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	if err := client.RemoveObject(ctx, p.config.Bucket, objectKey); err != nil {
		return fmt.Errorf("s3: failed to remove file after HTTP challenge: %w", err)
	}

	log.Info("s3: authz file removed", slog.String("bucket", p.config.Bucket), slog.String("object", objectKey))

	return nil
}

func (p *HTTPProvider) createS3Client() (*s3.Client, error) {
	clientCfg := s3.NewDefaultConfig()
	clientCfg.Endpoint = p.config.Endpoint
	clientCfg.AccessKey = p.config.AccessKey
	clientCfg.SecretKey = p.config.SecretKey
	clientCfg.SignatureVersion = p.config.SignatureVersion
	clientCfg.UsePathStyle = p.config.UsePathStyle
	clientCfg.Region = p.config.Region
	clientCfg.SkipTlsVerify = p.config.SkipTlsVerify

	client, err := s3.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
