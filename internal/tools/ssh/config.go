package ssh

const (
	defaultPort       int            = 22
	defaultAuthMethod AuthMethodType = AuthMethodTypeNone
	defaultUsername   string         = "root"
)

type ServerConfig struct {
	Host           string
	Port           int
	AuthMethod     AuthMethodType
	Username       string
	Password       string
	Key            string
	KeyPassphrase  string
	KeyCertificate string
}

type Config struct {
	ServerConfig
	JumpServers []ServerConfig
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:       defaultPort,
		AuthMethod: defaultAuthMethod,
		Username:   defaultUsername,
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		ServerConfig: *NewServerConfig(),
	}
}
