package teomakers

type Options struct {
	BaseEndpoint string
	ApiToken     string
}

type OptionsFunc func(*Options)

func WithChinaEndpoint() OptionsFunc {
	return func(o *Options) {
		o.BaseEndpoint = endpointChinaBaseURL
	}
}

func WithGlobalEndpoint() OptionsFunc {
	return func(o *Options) {
		o.BaseEndpoint = endpointGlobalBaseURL
	}
}

func WithApiToken(apiToken string) OptionsFunc {
	return func(o *Options) {
		o.ApiToken = apiToken
	}
}
