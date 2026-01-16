package s3

import (
	"bytes"
	"context"
	"errors"
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

const (
	SignatureV2 = "v2"
	SignatureV4 = "v4"
)

type Config struct {
	Endpoint         string
	AccessKey        string
	SecretKey        string
	SignatureVersion string // 默认值 "v4"
	UsePathStyle     bool
	Region           string
	SkipTlsVerify    bool
}

type Client struct {
	client *minio.Client
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("the configuration of S3 client is nil")
	}

	var clientCred *credentials.Credentials
	switch config.SignatureVersion {
	case "", SignatureV4:
		clientCred = credentials.NewStaticV4(config.AccessKey, config.SecretKey, "")
	case SignatureV2:
		clientCred = credentials.NewStaticV2(config.AccessKey, config.SecretKey, "")
	default:
		return nil, fmt.Errorf("unsupported S3 signature version: '%s'", config.SignatureVersion)
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

	if clientOpts.Secure && config.SkipTlsVerify {
		transport := xhttp.NewDefaultTransport()
		transport.TLSClientConfig = xtls.NewInsecureConfig()
		clientOpts.Transport = transport
	}

	client, err := minio.New(endpoint, clientOpts)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64) error {
	putOpts := minio.PutObjectOptions{
		DisableMultipart: true,
	}
	_, err := c.client.PutObject(ctx, bucket, key, reader, size, putOpts)
	if err != nil {
		return err
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
	err := c.client.RemoveObject(ctx, bucket, key, removeOpts)
	if err != nil {
		return err
	}

	return nil
}
