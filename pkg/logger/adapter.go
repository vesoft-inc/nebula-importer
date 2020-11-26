package logger

import (
	"fmt"
)

type NebulaLogger struct{}

func (l NebulaLogger) Info(msg string) {
	infoWithSkip(2, fmt.Sprintf("[nebula-clients] %s", msg))
}

func (l NebulaLogger) Warn(msg string) {
	warnWithSkip(2, fmt.Sprintf("[nebula-clients] %s", msg))
}

func (l NebulaLogger) Error(msg string) {
	errorWithSkip(2, fmt.Sprintf("[nebula-clients] %s", msg))
}

func (l NebulaLogger) Fatal(msg string) {
	fatalWithSkip(2, fmt.Sprintf("[nebula-clients] %s", msg))
}
