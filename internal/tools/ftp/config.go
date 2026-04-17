package ftp

const (
	defaultPort int = 21
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewDefaultConfig() *Config {
	return &Config{
		Port: defaultPort,
	}
}
