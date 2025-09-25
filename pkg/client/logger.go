package client

import (
	"log/slog"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

var _ nebula.Logger = nebulaLogger{}

type nebulaLogger struct {
	l *slog.Logger
}

func newNebulaLogger(l *slog.Logger) nebula.Logger {
	return nebulaLogger{
		l: l,
	}
}

//revive:disable:empty-lines

func (l nebulaLogger) Info(msg string)  { l.l.Info(msg) }
func (l nebulaLogger) Warn(msg string)  { l.l.Warn(msg) }
func (l nebulaLogger) Error(msg string) { l.l.Error(msg) }

//revive:enable:empty-lines
