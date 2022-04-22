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

// RunnerLogger TODO: Need to optimize it
type RunnerLogger struct {
	logger *log.Logger
}

func NewRunnerLogger(path string) *RunnerLogger {
	var w io.Writer
	if path != "" {
		file := base.MustCreateFile(path)
		w = io.MultiWriter(os.Stdout, file)
	} else {
		w = io.MultiWriter(os.Stdout)
	}
	logger := log.New(w, "", log.LstdFlags)
	r := new(RunnerLogger)
	r.logger = logger
	return r
}

func (r *RunnerLogger) Info(v ...interface{}) {
	r.infoWithSkip(2, fmt.Sprint(v...))
}

func (r *RunnerLogger) Infof(format string, v ...interface{}) {
	r.infoWithSkip(2, fmt.Sprintf(format, v...))
}

func (r *RunnerLogger) Warn(v ...interface{}) {
	r.warnWithSkip(2, fmt.Sprint(v...))
}

func (r *RunnerLogger) Warnf(format string, v ...interface{}) {
	r.warnWithSkip(2, fmt.Sprintf(format, v...))
}

func (r *RunnerLogger) Error(v ...interface{}) {
	r.errorWithSkip(2, fmt.Sprint(v...))
}

func (r *RunnerLogger) Errorf(format string, v ...interface{}) {
	r.errorWithSkip(2, fmt.Sprintf(format, v...))
}

func (r *RunnerLogger) Fatal(v ...interface{}) {
	r.fatalWithSkip(2, fmt.Sprint(v...))
}

func (r *RunnerLogger) Fatalf(format string, v ...interface{}) {
	r.fatalWithSkip(2, fmt.Sprintf(format, v...))
}

func (r *RunnerLogger) infoWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		r.logger.Printf("[INFO] %s:%d: %s", file, no, msg)
	} else {
		r.logger.Fatalf("Fail to get caller info of logger.Info")
	}
}

func (r *RunnerLogger) warnWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		r.logger.Printf("[WARN] %s:%d: %s", file, no, msg)
	} else {
		r.logger.Fatalf("Fail to get caller info of logger.Warn")
	}
}

func (r *RunnerLogger) errorWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		r.logger.Printf("[ERROR] %s:%d: %s", file, no, msg)
	} else {
		r.logger.Fatalf("Fail to get caller info of logger.Error")
	}
}

func (r *RunnerLogger) fatalWithSkip(skip int, msg string) {
	_, file, no, ok := runtime.Caller(skip)
	if ok {
		file = filepath.Base(file)
		r.logger.Fatalf("[FATAL] %s:%d: %s", file, no, msg)
	} else {
		r.logger.Fatalf("Fail to get caller info of logger.Fatal")
	}
}
