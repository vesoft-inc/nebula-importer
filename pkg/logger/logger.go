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
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[INFO] %s:%d: %s", file, no, fmt.Sprint(v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Info")
	}
}

func Infof(format string, v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[INFO] %s:%d: %s", file, no, fmt.Sprintf(format, v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Infof")
	}
}

func Warn(v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[WARN] %s:%d: %s", file, no, fmt.Sprint(v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Warn")
	}
}

func Warnf(format string, v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[WARN] %s:%d: %s", file, no, fmt.Sprintf(format, v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Warnf")
	}
}

func Error(v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[ERROR] %s:%d: %s", file, no, fmt.Sprint(v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Error")
	}
}

func Errorf(format string, v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Printf("[ERROR] %s:%d: %s", file, no, fmt.Sprintf(format, v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Errorf")
	}
}

func Fatal(v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Fatalf("[FATAL] %s:%d: %s", file, no, fmt.Sprint(v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Fatal")
	}
}

func Fatalf(format string, v ...interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		logger.Fatalf("[FATAL] %s:%d: %s", file, no, fmt.Sprintf(format, v...))
	} else {
		logger.Fatalf("Fail to get caller info of logger.Fatalf")
	}
}
