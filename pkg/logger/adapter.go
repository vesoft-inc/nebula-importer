package logger

import (
	"fmt"
)

type NebulaLogger struct {
	runnerLogger *RunnerLogger
}

func NewNebulaLogger(r *RunnerLogger) *NebulaLogger {
	n := new(NebulaLogger)
	n.runnerLogger = r
	return n
}

func (n NebulaLogger) Info(msg string) {
	n.runnerLogger.infoWithSkip(2, fmt.Sprintf("[nebula-go] %s", msg))
}

func (n NebulaLogger) Warn(msg string) {
	n.runnerLogger.warnWithSkip(2, fmt.Sprintf("[nebula-go] %s", msg))
}

func (n NebulaLogger) Error(msg string) {
	n.runnerLogger.errorWithSkip(2, fmt.Sprintf("[nebula-go] %s", msg))
}

func (n NebulaLogger) Fatal(msg string) {
	n.runnerLogger.fatalWithSkip(2, fmt.Sprintf("[nebula-go] %s", msg))
}
