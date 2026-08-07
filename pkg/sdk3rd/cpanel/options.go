package cpanel

type Options struct {
	Username string
	ApiToken string
}

type OptionsFunc func(*Options)

func WithUsername(username string) OptionsFunc {
	return func(o *Options) {
		o.Username = username
	}
}

func WithApiToken(apiToken string) OptionsFunc {
	return func(o *Options) {
		o.ApiToken = apiToken
	}
}
