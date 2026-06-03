package s3

import (
	"fmt"

	"github.com/certimate-go/certimate/pkg/core"
	"github.com/certimate-go/certimate/pkg/core/certifier/challengers/http01/s3/internal"
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

func NewChallenger(config *ChallengerConfig) (core.ACMEChallenger, error) {
	if config == nil {
		return nil, fmt.Errorf("the configuration of the acme challenge provider is nil")
	}

	providerConfig := internal.NewDefaultConfig()
	providerConfig.Endpoint = config.Endpoint
	providerConfig.AccessKey = config.AccessKey
	providerConfig.SecretKey = config.SecretKey
	providerConfig.SignatureVersion = config.SignatureVersion
	providerConfig.UsePathStyle = config.UsePathStyle
	providerConfig.Region = config.Region
	providerConfig.SkipTlsVerify = config.AllowInsecureConnections

	provider, err := internal.NewHTTPProviderConfig(providerConfig)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
