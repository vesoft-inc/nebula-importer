package logger

import (
	"log"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
)

var Log *log.Logger

func Init(path string) {
	file := base.MustCreateFile(path)
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
}
