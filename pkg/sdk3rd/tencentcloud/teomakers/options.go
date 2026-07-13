package teomakers

type Options struct {
	ApiToken string
}

type OptionsFunc func(*Options)

func WithApiToken(apiToken string) OptionsFunc {
	return func(o *Options) {
		o.ApiToken = apiToken
	}
}
