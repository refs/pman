package config

// Config determines behavior across the tool.
type Config struct {
	Hostname string
	Port     string
	File string
	KeepAlive bool
}

var (
	defaultHostname = "localhost"
	defaultPort     = "10666"
	defaultFile = "/var/tmp/.pman"
)

// NewConfig returns a new config with a set of defaults.
func NewConfig() *Config {
	return &Config{
		Hostname:  defaultHostname,
		Port:      defaultPort,
		File:      defaultFile,
		KeepAlive: false,
	}
}
