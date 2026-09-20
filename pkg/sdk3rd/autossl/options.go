package autossl

type Options struct {
	AccessKey string
	SecretKey string
}

type OptionsFunc func(*Options)

func WithAccessKey(accessKey string) OptionsFunc {
	return func(o *Options) {
		o.AccessKey = accessKey
	}
}

func WithSecretKey(secretKey string) OptionsFunc {
	return func(o *Options) {
		o.SecretKey = secretKey
	}
}
