package logger

type NebulaLogger struct{}

func (l NebulaLogger) Info(msg string) {
	Info(msg)
}

func (l NebulaLogger) Warn(msg string) {
	Warn(msg)
}

func (l NebulaLogger) Error(msg string) {
	Error(msg)
}

func (l NebulaLogger) Fatal(msg string) {
	Fatal(msg)
}
