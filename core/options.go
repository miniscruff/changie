package core

type (
	// Option to the core package utilities.
	Option func(*coreOptions)

	coreOptions struct {
		skipPrereleases bool
	}
)

func optionsFromDefault(opts ...Option) *coreOptions {
	options := &coreOptions{}

	for _, apply := range opts {
		apply(options)
	}

	return options
}

// WithSkipPrereleases includes prereleases when determining the latest version.
func WithSkipPrereleases(skipPrereleases bool) Option {
	return func(o *coreOptions) {
		o.skipPrereleases = skipPrereleases
	}
}
