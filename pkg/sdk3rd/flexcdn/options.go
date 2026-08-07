package flexcdn

type Options struct {
	Role        string
	AccessKeyId string
	AccessKey   string
}

type OptionsFunc func(*Options)

func WithRole(role string) OptionsFunc {
	return func(o *Options) {
		o.Role = role
	}
}

func WithAccessKey(accessKeyId, accessKey string) OptionsFunc {
	return func(o *Options) {
		o.AccessKeyId = accessKeyId
		o.AccessKey = accessKey
	}
}
