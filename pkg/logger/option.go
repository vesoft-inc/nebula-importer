package logger

import "time"

var defaultOptions = options{
	level:      InfoLevel,
	console:    true,
	timeLayout: time.RFC3339,
}

type (
	options struct {
		level      Level
		fields     Fields
		console    bool
		timeLayout string
		files      []string
	}
	Option func(*options)
)

func WithLevel(lvl Level) Option {
	return func(o *options) {
		o.level = lvl
	}
}

func WithLevelText(text string) Option {
	return func(o *options) {
		WithLevel(ParseLevel(text))(o)
	}
}

func WithFields(fields ...Field) Option {
	return func(o *options) {
		o.fields = fields
	}
}

func WithConsole(console bool) Option {
	return func(o *options) {
		o.console = console
	}
}

func WithTimeLayout(layout string) Option {
	return func(o *options) {
		o.timeLayout = layout
	}
}

func WithFiles(files ...string) Option {
	return func(o *options) {
		o.files = files
	}
}
