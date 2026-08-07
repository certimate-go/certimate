package matrix

type Options struct {
	UserId      string
	AccessToken string
}

type OptionsFunc func(*Options)

func WithUserId(userId string) OptionsFunc {
	return func(o *Options) {
		o.UserId = userId
	}
}

func WithAccessToken(accessToken string) OptionsFunc {
	return func(o *Options) {
		o.AccessToken = accessToken
	}
}
