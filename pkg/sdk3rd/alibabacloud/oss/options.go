package oss

type Options struct {
	AccessKeyId     string
	AccessKeySecret string
	Region          string
	Bucket          string
}

type OptionsFunc func(*Options)

func WithAkSk(ak, sk string) OptionsFunc {
	return func(o *Options) {
		o.AccessKeyId = ak
		o.AccessKeySecret = sk
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
