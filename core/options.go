package core

type (
	// Option to the core package utilities.
	Option func(*coreOptions)

	coreOptions struct {
		forcePatch bool
	}
)

func optionsFromDefault(opts ...Option) *coreOptions {
	options := &coreOptions{}

	for _, apply := range opts {
		apply(options)
	}

	return options
}

// WithForcePatch forces the increment of the patch version when getting the next release.
func WithForcePatch(force bool) Option {
	return func(o *coreOptions) {
		o.forcePatch = force
	}
}
