package inlinecode

import "github.com/a-h/templ"

type config struct {
	text      string
	rootClass string
	rootAttrs templ.Attributes
}

// Option configures inline code.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(cfg *config) {
	fn(cfg)
}

// WithRootClass appends CSS classes to the code element.
func WithRootClass(class string) Option {
	return optionFunc(func(cfg *config) {
		cfg.rootClass = class
	})
}

// WithRootAttrs adds arbitrary attributes to the code element.
func WithRootAttrs(attrs templ.Attributes) Option {
	return optionFunc(func(cfg *config) {
		cfg.rootAttrs = attrs
	})
}

func newConfig(text string, options []Option) config {
	cfg := config{text: text}
	for _, option := range options {
		if option != nil {
			option.apply(&cfg)
		}
	}
	return cfg
}

func (cfg config) rootClasses() string {
	classes := "rounded-radius bg-surface-alt px-1.5 py-0.5 font-mono text-[0.875em] text-on-surface-strong dark:bg-surface-dark-alt dark:text-on-surface-dark-strong"
	if cfg.rootClass != "" {
		classes += " " + cfg.rootClass
	}
	return classes
}
