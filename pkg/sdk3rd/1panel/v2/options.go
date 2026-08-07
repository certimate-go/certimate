package v2

type Options struct {
	ApiKey      string
	CurrentNode string
}

type OptionsFunc func(*Options)

func WithApiKey(apiKey string) OptionsFunc {
	return func(o *Options) {
		o.ApiKey = apiKey
	}
}

func WithNode(node string) OptionsFunc {
	if node == "" {
		node = "local"
	}

	return func(o *Options) {
		o.CurrentNode = node
	}
}
