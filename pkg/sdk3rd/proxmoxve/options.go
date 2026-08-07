package proxmoxve

type Options struct {
	TokenId     string
	TokenSecret string
}

type OptionsFunc func(*Options)

func WithApiToken(tokenId string, tokenSecret string) OptionsFunc {
	return func(o *Options) {
		o.TokenId = tokenId
		o.TokenSecret = tokenSecret
	}
}
