package cdnfly

type Options struct {
	ApiKey    string
	ApiSecret string
}

type OptionsFunc func(*Options)

func WithApiKey(apiKey, apiSecret string) OptionsFunc {
	return func(o *Options) {
		o.ApiKey = apiKey
		o.ApiSecret = apiSecret
	}
}
