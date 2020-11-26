package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
)

var logger *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

func Init(path string) {
	file := base.MustCreateFile(path)
	w := io.MultiWriter(os.Stdout, file)
	logger = log.New(w, "", log.LstdFlags)
}

func Info(v ...interface{}) {
	infoWithSkip(2, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	infoWithSkip(2, fmt.Sprintf(format, fmt.Sprintf(format, v...)))
}

func Warn(v ...interface{}) {
	warnWithSkip(2, fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	warnWithSkip(2, fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	errorWithSkip(2, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	errorWithSkip(2, fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	fatalWithSkip(2, fmt.Sprint(v...))
}

func Fatalf(format string, v ...interface{}) {
	fatalWithSkip(2, fmt.Sprintf(format, v...))
}

func infoWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[INFO] %s:%d: %s", file, no, msg)
	} else {
		logger.Fatalf("Fail to get caller info of logger.Info")
	}
}

func warnWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[WARN] %s:%d: %s", file, no, msg)
	} else {
		logger.Fatalf("Fail to get caller info of logger.Warn")
	}
}

func errorWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[ERROR] %s:%d: %s", file, no, msg)
	} else {
		logger.Fatalf("Fail to get caller info of logger.Error")
	}
}

func fatalWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		logger.Fatalf("[FATAL] %s:%d: %s", file, no, msg)
	} else {
		logger.Fatalf("Fail to get caller info of logger.Fatal")
	}
}
