package errhandler

import (
	"log"
	"strings"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/csv"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

type ErrData struct {
	Error error
	Data  base.Data
	Done  bool
}

type ErrorWriter interface {
	GetErrorChan() chan ErrData
	InitFile(config.File)
}

func New(errCh <-chan ErrData, failCh chan<- stats.Stats) ErrorWriter {
	switch strings.ToUpper(file.Type) {
	case "CSV":
		return &csv.CSVErrWriter{
			ErrCh:  errCh,
			FailCh: failCh,
		}
	default:
		log.Fatalf("Wrong file type: %s", file.Type)
		return nil
	}
}
