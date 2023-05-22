package logger

type (
	Logger interface {
		SkipCaller(skip int) Logger

		With(fields ...Field) Logger
		WithError(err error) Logger

		Debug(msg string, fields ...Field)
		Info(msg string, fields ...Field)
		Warn(msg string, fields ...Field)
		Error(msg string, fields ...Field)
		Panic(msg string, fields ...Field)
		Fatal(msg string, fields ...Field)

		Sync() error
		Close() error
	}
)

func New(opts ...Option) (Logger, error) {
	o := defaultOptions

	for _, opt := range opts {
		opt(&o)
	}

	l, err := newZapLogger(&o)
	if err != nil {
		return nil, err
	}
	return l, nil
}
