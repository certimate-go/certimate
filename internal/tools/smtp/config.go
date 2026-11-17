package smtp

const (
	defaultPort int = 25
)

type Config struct {
	Host          string
	Port          int
	Username      string
	Password      string
	UseSsl        bool
	SkipTlsVerify bool
}

func NewDefaultConfig() *Config {
	return &Config{
		Port: defaultPort,
	}
}
