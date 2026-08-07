package btpanel

type Options struct {
	ApiKey string
}

type OptionsFunc func(*Options)

func WithApiKey(apiKey string) OptionsFunc {
	return func(o *Options) {
		o.ApiKey = apiKey
	}
}
