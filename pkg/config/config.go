package config

import (
	"io/ioutil"
	"log"
)

// Config determines behavior across the tool.
type Config struct {
	// Hostname where the runtime is running. When using PMAN in cli mode, it determines where the host runtime is.
	// Default is localhost.
	Hostname string

	// Port configures the port where a runtime is available. It defaults to 10666.
	Port     string

	// File configures where Pman's database of PID lives.
	File string

	// KeepAlive configures if restart attempts are made if the process supervised terminates. Default is false.
	KeepAlive bool
}

// NewConfig returns a new config with a set of defaults.
func NewConfig() *Config {
	f, err := ioutil.TempFile("", "*")
	if err != nil {
		log.Fatal(err)
	}
	mustWrite(f.Name(), []byte("{}"))

	return &Config{
		Hostname:  defaultHostname,
		Port:      defaultPort,
		File:      f.Name(),
		KeepAlive: false,
	}
}
