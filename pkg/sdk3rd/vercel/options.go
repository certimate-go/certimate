package vercel

type Options struct {
	ApiToken string
	TeamId   string
}

type OptionsFunc func(*Options)

func WithApiToken(apiToken string) OptionsFunc {
	return func(o *Options) {
		o.ApiToken = apiToken
	}
}

func WithTeamId(teamId string) OptionsFunc {
	return func(o *Options) {
		o.TeamId = teamId
	}
}
