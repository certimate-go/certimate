package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-acme/lego/v4/challenge/http01"

	"github.com/certimate-go/certimate/internal/tools/s3"
	"github.com/certimate-go/certimate/pkg/core/certifier"
)

type ChallengerConfig struct {
	// S3 Endpoint。
	Endpoint string `json:"endpoint"`
	// S3 AccessKey。
	AccessKey string `json:"accessKey"`
	// S3 SecretKey。
	SecretKey string `json:"secretKey"`
	// S3 签名版本。
	// 可取值 "v2"、"v4"。
	// 零值时默认值 "v4"。
	SignatureVersion string `json:"signatureVersion,omitempty"`
	// 是否使用路径风格。
	UsePathStyle bool `json:"usePathStyle,omitempty"`
	// 存储区域。
	Region string `json:"region"`
	// 存储桶名。
	Bucket string `json:"bucket"`
	// 是否允许不安全的连接。
	AllowInsecureConnections bool `json:"allowInsecureConnections,omitempty"`
}

func NewChallenger(config *ChallengerConfig) (certifier.ACMEChallenger, error) {
	if config == nil {
		return nil, errors.New("the configuration of the acme challenge provider is nil")
	}

	client, err := s3.NewClient(&s3.Config{
		Endpoint:         config.Endpoint,
		AccessKey:        config.AccessKey,
		SecretKey:        config.SecretKey,
		SignatureVersion: config.SignatureVersion,
		UsePathStyle:     config.UsePathStyle,
		Region:           config.Region,
		SkipTlsVerify:    config.AllowInsecureConnections,
	})
	if err != nil {
		return nil, fmt.Errorf("s3: failed to create s3 client: %w", err)
	}

	provider := &provider{client: client, bucket: config.Bucket}
	return provider, nil
}

type provider struct {
	client *s3.Client
	bucket string
}

func (p *provider) Present(domain, token, keyAuth string) error {
	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	if err := p.client.PutObjectString(context.Background(), p.bucket, objectKey, keyAuth); err != nil {
		return fmt.Errorf("s3: failed to upload token to s3: %w", err)
	}

	return nil
}

func (p *provider) CleanUp(domain, token, keyAuth string) error {
	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	if err := p.client.RemoveObject(context.Background(), p.bucket, objectKey); err != nil {
		return fmt.Errorf("s3: could not remove file in s3 bucket after HTTP challenge: %w", err)
	}

	return nil
}
