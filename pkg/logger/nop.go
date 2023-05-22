package logger

var NopLogger Logger = nopLogger{}

type nopLogger struct{}

//revive:disable:empty-lines

func (l nopLogger) SkipCaller(int) Logger  { return l }
func (l nopLogger) With(...Field) Logger   { return l }
func (l nopLogger) WithError(error) Logger { return l }
func (nopLogger) Debug(string, ...Field)   {}
func (nopLogger) Info(string, ...Field)    {}
func (nopLogger) Warn(string, ...Field)    {}
func (nopLogger) Error(string, ...Field)   {}
func (nopLogger) Panic(string, ...Field)   {}
func (nopLogger) Fatal(string, ...Field)   {}
func (nopLogger) Sync() error              { return nil }
func (nopLogger) Close() error             { return nil }

//revive:enable:empty-lines
