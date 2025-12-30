package s3

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/lo"

	"github.com/certimate-go/certimate/pkg/core/certifier"
	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
	xtls "github.com/certimate-go/certimate/pkg/utils/tls"
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

	var clientCred *credentials.Credentials
	switch config.SignatureVersion {
	case "", "v4":
		clientCred = credentials.NewStaticV4(config.AccessKey, config.SecretKey, "")
	case "v2":
		clientCred = credentials.NewStaticV2(config.AccessKey, config.SecretKey, "")
	default:
		return nil, fmt.Errorf("unsupported s3 signature version: '%s'", config.SignatureVersion)
	}

	var clientOpts *minio.Options
	clientOpts = &minio.Options{
		Creds:        clientCred,
		Region:       config.Region,
		BucketLookup: lo.If(config.UsePathStyle, minio.BucketLookupPath).Else(minio.BucketLookupDNS),
	}

	var endpoint string
	if config.Endpoint != "" {
		reScheme := regexp.MustCompile("^([^:]+)://")
		if reScheme.MatchString(config.Endpoint) {
			temp := strings.Split(config.Endpoint, "://")
			scheme := temp[0]
			endpoint = temp[1]
			clientOpts.Secure = strings.EqualFold(scheme, "https")
		} else {
			endpoint = config.Endpoint
			clientOpts.Secure = true
		}
	}

	if clientOpts.Secure && config.AllowInsecureConnections {
		transport := xhttp.NewDefaultTransport()
		transport.DisableKeepAlives = true
		transport.TLSClientConfig = xtls.NewInsecureConfig()
		clientOpts.Transport = transport
	}

	client, err := minio.New(endpoint, clientOpts)
	if err != nil {
		return nil, err
	}

	provider := &provider{client: client, bucket: config.Bucket}
	return provider, nil
}

type provider struct {
	client *minio.Client
	bucket string
}

func (p *provider) Present(domain, token, keyAuth string) error {
	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	putOpts := minio.PutObjectOptions{
		DisableMultipart: true,
	}
	reader := strings.NewReader(keyAuth)
	_, err := p.client.PutObject(context.Background(), p.bucket, objectKey, reader, reader.Size(), putOpts)
	if err != nil {
		return fmt.Errorf("s3: failed to upload token to s3: %w", err)
	}

	return nil
}

func (p *provider) CleanUp(domain, token, keyAuth string) error {
	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	removeOpts := minio.RemoveObjectOptions{}
	err := p.client.RemoveObject(context.Background(), p.bucket, objectKey, removeOpts)
	if err != nil {
		return fmt.Errorf("s3: could not remove file in s3 bucket after HTTP challenge: %w", err)
	}

	return nil
}
