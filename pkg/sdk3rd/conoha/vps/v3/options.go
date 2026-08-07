package v3

type Options struct {
	UserId       string
	UserName     string
	UserPassword string
	TenantId     string
	TenantName   string
}

type OptionsFunc func(*Options)

func WithUserId(userId string) OptionsFunc {
	return func(o *Options) {
		o.UserId = userId
	}
}

func WithUserName(userName string) OptionsFunc {
	return func(o *Options) {
		o.UserName = userName
	}
}

func WithUserPassword(userPassword string) OptionsFunc {
	return func(o *Options) {
		o.UserPassword = userPassword
	}
}

func WithTenantId(tenantId string) OptionsFunc {
	return func(o *Options) {
		o.TenantId = tenantId
	}
}

func WithTenantName(tenantName string) OptionsFunc {
	return func(o *Options) {
		o.TenantName = tenantName
	}
}
