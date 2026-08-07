package ratpanel

type Options struct {
	AccessTokenId int64
	AccessToken   string
}

type OptionsFunc func(*Options)

func WithAccessToken(accessTokenId int64, accessToken string) OptionsFunc {
	return func(o *Options) {
		o.AccessTokenId = accessTokenId
		o.AccessToken = accessToken
	}
}
