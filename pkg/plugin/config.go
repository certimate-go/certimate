package plugin

import "time"

const (
	defaultMinPort       = 10000
	defaultMaxPort       = 25000
	defaultStartTimeout  = 10 * time.Second
	defaultDeployTimeout = 30 * time.Minute
	defaultMaxLogFrames  = 1000
)

type PluginConfig struct {
	PluginDir     string
	MinPort       uint
	MaxPort       uint
	StartTimeout  time.Duration
	DeployTimeout time.Duration
	MaxLogFrames  int
	LogLevel      string
	CoreVersion   string
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
	if c.DeployTimeout == 0 {
		c.DeployTimeout = defaultDeployTimeout
	}
	if c.MaxLogFrames == 0 {
		c.MaxLogFrames = defaultMaxLogFrames
	}
	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}
}
