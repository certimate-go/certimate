package common

type Options struct {
	AccessKeyId     string
	SecretAccessKey string
}

type OptionsFunc func(*Options)

func WithAkSk(ak, sk string) OptionsFunc {
	return func(o *Options) {
		o.AccessKeyId = ak
		o.SecretAccessKey = sk
	}
}
