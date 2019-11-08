package logger

import (
	"io"
	"log"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
)

var Log *log.Logger

func Init(path string) {
	file := base.MustCreateFile(path)
	w := io.MultiWriter(os.Stdout, file)
	Log = log.New(w, "", log.LstdFlags|log.Lshortfile)
}
