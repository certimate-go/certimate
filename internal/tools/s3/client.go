package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/lo"

	xhttp "github.com/certimate-go/certimate/pkg/utils/http"
	xtls "github.com/certimate-go/certimate/pkg/utils/tls"
)

type Client struct {
	cli *minio.Client
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of S3 client is nil")
	}

	client, err := createS3Client(config)
	if err != nil {
		return nil, err
	}

	return &Client{cli: client}, nil
}

func (c *Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	putOpts := minio.PutObjectOptions{
		DisableMultipart: true,
	}
	_, err := c.cli.PutObject(ctx, bucket, key, reader, size, putOpts)
	if err != nil {
		return fmt.Errorf("s3: failed to put object: %w", err)
	}

	return nil
}

func (c *Client) PutObjectString(ctx context.Context, bucket, key string, data string) error {
	reader := strings.NewReader(data)
	return c.PutObject(ctx, bucket, key, reader, reader.Size())
}

func (c *Client) PutObjectBytes(ctx context.Context, bucket, key string, data []byte) error {
	reader := bytes.NewReader(data)
	return c.PutObject(ctx, bucket, key, reader, reader.Size())
}

func (c *Client) RemoveObject(ctx context.Context, bucket, key string) error {
	removeOpts := minio.RemoveObjectOptions{}
	err := c.cli.RemoveObject(ctx, bucket, key, removeOpts)
	if err != nil {
		return fmt.Errorf("s3: failed to remove object: %w", err)
	}

	return nil
}

func createS3Client(config *Config) (*minio.Client, error) {
	var clientCred *credentials.Credentials
	switch config.SignatureVersion {
	case "", SignatureV4:
		clientCred = credentials.NewStaticV4(config.AccessKey, config.SecretKey, "")
	case SignatureV2:
		clientCred = credentials.NewStaticV2(config.AccessKey, config.SecretKey, "")
	default:
		return nil, fmt.Errorf("s3: unsupported signature version: '%s'", config.SignatureVersion)
	}

	endpoint, secure := resolveEndpoint(config.Endpoint)
	clientOpts := &minio.Options{
		Creds:        clientCred,
		Region:       config.Region,
		BucketLookup: lo.If(config.UsePathStyle, minio.BucketLookupPath).Else(minio.BucketLookupDNS),
		Secure:       secure,
	}

	if secure && config.SkipTlsVerify {
		transport := xhttp.NewDefaultTransport()
		transport.TLSClientConfig = xtls.NewInsecureConfig()
		clientOpts.Transport = transport
	}

	client, err := minio.New(endpoint, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("s3: %w", err)
	}

	return client, nil
}

func resolveEndpoint(endpoint string) (string, bool) {
	var secure bool
	var result string

	reScheme := regexp.MustCompile(`^([^:]+)://`)
	if reScheme.MatchString(endpoint) {
		temp := strings.Split(endpoint, "://")
		scheme := temp[0]
		result = temp[1]
		secure = strings.EqualFold(scheme, "https")
	} else {
		result = endpoint
		secure = true
	}

	return result, secure
}
