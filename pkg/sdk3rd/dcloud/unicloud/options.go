package unicloud

type Options struct {
	Username string
	Password string
}

type OptionsFunc func(*Options)

func WithLogins(username, password string) OptionsFunc {
	return func(o *Options) {
		o.Username = username
		o.Password = password
	}
}
