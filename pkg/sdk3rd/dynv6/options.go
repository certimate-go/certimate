package dynv6

type Options struct {
	HttpToken string
}

type OptionsFunc func(*Options)

func WithHttpToken(httpToken string) OptionsFunc {
	return func(o *Options) {
		o.HttpToken = httpToken
	}
}
