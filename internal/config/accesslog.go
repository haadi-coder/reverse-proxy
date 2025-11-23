package config

type AccessLogConfig struct {
	Format string `yaml:"format,omitempty"`
}

func (c *AccessLogConfig) applyDefaults() {
	if c != nil {
		if c.Format == "" {
			c.Format = "common"
		}
	}
}
