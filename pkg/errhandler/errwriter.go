package errhandler

import (
	"log"
	"strings"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/csv"
)

type ErrorWriter interface {
	SetupErrorHandler()
}

func New(file config.File, errCh <-chan base.ErrData, failCh chan<- bool) ErrorWriter {
	switch strings.ToUpper(file.Type) {
	case "CSV":
		return &csv.CSVErrWriter{
			ErrConf: file.Error,
			ErrCh:   errCh,
			FailCh:  failCh,
		}
	default:
		log.Fatalf("Wrong file type: %s", file.Type)
		return nil
	}
}
