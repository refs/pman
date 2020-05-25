package config

// Config determines behavior across the tool.
type Config struct {
	Hostname string
	Port     string
}

var (
	defaultHostname = "localhost"
	defaultPort     = "10666"
)

// NewConfig returns a new config with a set of defaults.
func NewConfig() *Config {
	return &Config{
		Hostname: defaultHostname,
		Port:     defaultPort,
	}
}
