package digitalocean

type Options struct {
	AccessToken string
}

type OptionsFunc func(*Options)

func WithAccessToken(accessToken string) OptionsFunc {
	return func(o *Options) {
		o.AccessToken = accessToken
	}
}
