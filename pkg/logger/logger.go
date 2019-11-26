package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
)

var logger *log.Logger

func Init(path string) {
	file := base.MustCreateFile(path)
	w := io.MultiWriter(os.Stdout, file)
	logger = log.New(w, "", log.LstdFlags|log.Lshortfile)
}

func Info(v ...interface{}) {
	logger.Printf("[INFO] %s", fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	logger.Printf("[INFO] %s", fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	logger.Printf("[WARN] %s", fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	logger.Printf("[WARN] %s", fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	logger.Printf("[ERROR] %s", fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	logger.Printf("[ERROR] %s", fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	logger.Fatal(fmt.Sprint(v...))
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatal(fmt.Errorf(format, v...))
}
