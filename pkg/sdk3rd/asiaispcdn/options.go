package asiaispcdn

type Options struct {
	AccessKeyId     string
	AccessKeySecret string
}

type OptionsFunc func(*Options)

func WithAkSk(ak, sk string) OptionsFunc {
	return func(o *Options) {
		o.AccessKeyId = ak
		o.AccessKeySecret = sk
	}
}
