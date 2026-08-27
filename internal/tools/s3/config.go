package s3

const (
	SignatureV2 = "v2"
	SignatureV4 = "v4"
)

const (
	defaultSignatureVersion = SignatureV4
)

type Config struct {
	Endpoint         string
	AccessKey        string
	SecretKey        string
	SignatureVersion string
	UsePathStyle     bool
	Region           string
	SkipTlsVerify    bool
}

func NewDefaultConfig() *Config {
	return &Config{
		SignatureVersion: defaultSignatureVersion,
	}
}
