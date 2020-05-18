package controller

// Options are the configurable options for a Controller.
type Options struct {
	Bin string
}

// Option represents an option.
type Option func(o Options)

// NewOptions returns a new Options struct.
func NewOptions() Options {
	return Options{}
}

// WithBinary set binary on Options.
func WithBinary(bin string) Option {
	return func(o Options) {
		o.Bin = bin
	}
}
