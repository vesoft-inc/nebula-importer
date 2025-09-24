package reader

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
)

const (
	DefaultBatchSize = 128
)

type (
	Option func(*options)

	options struct {
		batch  int
		logger logger.Logger
	}
)

func WithBatch(batch int) Option {
	return func(m *options) {
		m.batch = batch
	}
}

func WithLogger(l logger.Logger) Option {
	return func(m *options) {
		m.logger = l
	}
}

func newOptions(opts ...Option) *options {
	defaultOptions := &options{
		batch: DefaultBatchSize,
	}

	defaultOptions.withOptions(opts...)

	if defaultOptions.batch <= 0 {
		defaultOptions.batch = DefaultBatchSize
	}

	return defaultOptions
}

func (o *options) withOptions(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}

	if o.logger == nil {
		o.logger = logger.NopLogger
	}
}
