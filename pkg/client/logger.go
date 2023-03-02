package client

import (
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
)

var _ nebula.Logger = nebulaLogger{}

type nebulaLogger struct {
	l logger.Logger
}

func newNebulaLogger(l logger.Logger) nebula.Logger {
	return nebulaLogger{
		l: l,
	}
}

//revive:disable:empty-lines

func (l nebulaLogger) Info(msg string)  { l.l.Info(msg) }
func (l nebulaLogger) Warn(msg string)  { l.l.Warn(msg) }
func (l nebulaLogger) Error(msg string) { l.l.Error(msg) }
func (l nebulaLogger) Fatal(msg string) { l.l.Fatal(msg) }

//revive:enable:empty-lines
