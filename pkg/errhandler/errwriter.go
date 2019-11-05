package errhandler

import (
	"log"
	"strings"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/csv"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

type ErrorWriter interface {
	InitFile(config.File, int)
}

func New(file config.File, errCh <-chan base.ErrData, failCh chan<- stats.Stats) ErrorWriter {
	switch strings.ToUpper(file.Type) {
	case "CSV":
		w := csv.CSVErrWriter{
			ErrCh:  errCh,
			FailCh: failCh,
		}
		return &w
	default:
		log.Fatalf("Wrong file type: %s", file.Type)
		return nil
	}
}
