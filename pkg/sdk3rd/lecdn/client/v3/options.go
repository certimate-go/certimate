package v3

type Options struct {
	Username string
	Password string
	ApiKey   string
}

type OptionsFunc func(*Options)

func WithLogins(username, password string) OptionsFunc {
	return func(o *Options) {
		o.Username = username
		o.Password = password
		o.ApiKey = ""
	}
}

func WithApiKey(apiKey string) OptionsFunc {
	return func(o *Options) {
		o.ApiKey = apiKey
		o.Username = ""
		o.Password = ""
	}
}
