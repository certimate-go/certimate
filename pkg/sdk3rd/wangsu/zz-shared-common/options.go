package common

type Options struct {
	AccessKey string
	SecretKey string
}

type OptionsFunc func(*Options)

func WithAkSk(ak, sk string) OptionsFunc {
	return func(o *Options) {
		o.AccessKey = ak
		o.SecretKey = sk
	}
}
