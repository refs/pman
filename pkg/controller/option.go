package controller

// Options are the configurable options for a Controller.
type Options struct {
	Bin  string
	File string
	Restart bool
	Grace int
}

// Option represents an option.
type Option func(o *Options)

// NewOptions returns a new Options struct.
func NewOptions() Options {
	return Options{}
}

// WithBinary set binary on Options.
func WithBinary(bin string) Option {
	return func(o *Options) {
		o.Bin = bin
	}
}

// WithFile set the db file to store a list of managed processes PID.
func WithFile(file string) Option {
	return func(o *Options) {
		o.File = file
	}
}

// WithRestart sets restart, which control whether a controller restart killed processes.
func WithRestart(r bool) Option {
	return func(o *Options) {
		o.Restart = r
	}
}

// WithGrace sets restart, which control whether a controller restart killed processes.
func WithGrace(g int) Option {
	return func(o *Options) {
		o.Grace = g
	}
}
