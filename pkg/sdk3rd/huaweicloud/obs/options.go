package obs

type Options struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string
	Bucket          string
}

type OptionsFunc func(*Options)

func WithAkSk(ak, sk string) OptionsFunc {
	return func(o *Options) {
		o.AccessKeyId = ak
		o.SecretAccessKey = sk
	}
}

func WithRegion(region string) OptionsFunc {
	return func(o *Options) {
		o.Region = region
	}
}

func WithBucket(bucket string) OptionsFunc {
	return func(o *Options) {
		o.Bucket = bucket
	}
}
