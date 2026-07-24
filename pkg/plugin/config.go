package plugin

import "time"

const (
	defaultMinPort      = 10000
	defaultMaxPort      = 25000
	defaultStartTimeout = 10 * time.Second
)

type PluginConfig struct {
	PluginDir    string
	MinPort      uint
	MaxPort      uint
	StartTimeout time.Duration
	LogLevel     string
	CoreVersion  string
}

func (c *PluginConfig) defaults() {
	if c.MinPort == 0 {
		c.MinPort = defaultMinPort
	}
	if c.MaxPort == 0 {
		c.MaxPort = defaultMaxPort
	}
	if c.MaxPort < c.MinPort {
		c.MinPort, c.MaxPort = c.MaxPort, c.MinPort
	}
	if c.StartTimeout == 0 {
		c.StartTimeout = defaultStartTimeout
	}
	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}
}
