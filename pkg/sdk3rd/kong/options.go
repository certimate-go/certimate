package kong

type Options struct {
	ApiToken  string
	Workspace string
}

type OptionsFunc func(*Options)

func WithApiToken(apiToken string) OptionsFunc {
	return func(o *Options) {
		o.ApiToken = apiToken
	}
}

func WithWorkspace(workspace string) OptionsFunc {
	return func(o *Options) {
		o.Workspace = workspace
	}
}
