package controller

import "github.com/refs/pman/pkg/config"

// Options are the configurable options for a Controller.
type Options struct {
	Bin  string
	Restart bool
	Config *config.Config
}

// Option represents an option.
type Option func(o *Options)

// NewOptions returns a new Options struct.
func NewOptions() Options {
	return Options{}
}

// WithConfig sets Controller config.
func WithConfig(cfg *config.Config) Option {
	return func(o *Options) {
		o.Config = cfg
	}
}
