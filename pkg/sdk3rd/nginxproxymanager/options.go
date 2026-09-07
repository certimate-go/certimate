package nginxproxymanager

type Options struct {
	Username string
	Password string
	JwtToken string
}

type OptionsFunc func(*Options)

func WithLogins(username, password string) OptionsFunc {
	return func(o *Options) {
		o.Username = username
		o.Password = password
		o.JwtToken = ""
	}
}

func WithJwtToken(jwtToken string) OptionsFunc {
	return func(o *Options) {
		o.JwtToken = jwtToken
		o.Username = ""
		o.Password = ""
	}
}
