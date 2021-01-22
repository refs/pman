package service

// Log configures a structure logger.
type Log struct {
	Pretty bool
}

// Options are the configurable options for a Service.
type Options struct {
	Log *Log
}

// Option represents an option.
type Option func(o *Options)

// NewOptions returns a new Options struct.
func NewOptions() *Options {
	return &Options{
		Log: &Log{},
	}
}

// WithLogPretty sets Controller config.
func WithLogPretty(pretty bool) Option {
	return func(o *Options) {
		o.Log.Pretty = pretty
	}
}
