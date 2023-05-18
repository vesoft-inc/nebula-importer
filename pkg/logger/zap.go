package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	l       *zap.Logger
	cleanup func()
}

func newZapLogger(o *options) (*zapLogger, error) {
	l := &zapLogger{}

	atomicLevel := zap.NewAtomicLevelAt(toZapLevel(o.level))

	var cores []zapcore.Core
	encoderCfg := zap.NewProductionEncoderConfig()
	if o.timeLayout != "" {
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(o.timeLayout)
	}
	if o.console {
		cores = append(cores,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				zapcore.Lock(os.Stdout),
				atomicLevel,
			),
		)
	}
	if len(o.files) > 0 {
		sink, cleanup, err := zap.Open(o.files...)
		if err != nil {
			return nil, err
		}
		l.cleanup = cleanup
		cores = append(cores,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				sink,
				atomicLevel,
			),
		)
	}

	l.l = zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(toZapFields(o.fields...)...),
	)

	return l, nil
}

func (l *zapLogger) SkipCaller(skip int) Logger {
	cpy := l.clone()
	cpy.l = cpy.l.WithOptions(zap.AddCallerSkip(skip))
	return cpy
}

func (l *zapLogger) With(fields ...Field) Logger {
	cpy := l.clone()
	cpy.l = cpy.l.With(toZapFields(fields...)...)
	return cpy
}

func (l *zapLogger) WithError(err error) Logger {
	cpy := l.clone()
	cpy.l = cpy.l.With(zap.Error(err))
	return cpy
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.l.Debug(msg, toZapFields(fields...)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.l.Info(msg, toZapFields(fields...)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, toZapFields(fields...)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.l.Error(msg, toZapFields(fields...)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, toZapFields(fields...)...)
}

func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.l.Panic(msg, toZapFields(fields...)...)
}

func (l *zapLogger) Sync() error {
	return l.l.Sync()
}

func (l *zapLogger) Close() error {
	if l.cleanup != nil {
		defer l.cleanup()
	}
	//revive:disable-next-line:if-return
	if err := l.Sync(); err != nil {
		return err
	}
	return nil
}

func (l *zapLogger) clone() *zapLogger {
	cpy := *l
	return &cpy
}

func toZapFields(fields ...Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

func toZapLevel(lvl Level) zapcore.Level {
	switch lvl {
	case DebugLevel:
		return zap.DebugLevel
	case InfoLevel:
		return zap.InfoLevel
	case WarnLevel:
		return zap.WarnLevel
	case ErrorLevel:
		return zap.ErrorLevel
	case PanicLevel:
		return zap.PanicLevel
	case FatalLevel:
		return zap.FatalLevel
	}
	return zap.InfoLevel
}
