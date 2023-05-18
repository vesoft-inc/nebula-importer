package logger

import "strings"

type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

var (
	levelMap = map[Level]string{
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
		PanicLevel: "PANIC",
		FatalLevel: "FATAL",
	}
	levelRevMap = map[string]Level{
		"DEBUG": DebugLevel,
		"INFO":  InfoLevel,
		"WARN":  WarnLevel,
		"ERROR": ErrorLevel,
		"PANIC": PanicLevel,
		"FATAL": FatalLevel,
	}
)

func ParseLevel(text string) Level {
	level, ok := levelRevMap[strings.ToUpper(text)]
	if !ok {
		level = InfoLevel
	}
	return level
}

func (lvl Level) String() string {
	return levelMap[lvl]
}
