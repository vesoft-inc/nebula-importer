package reader

import (
	"log"
	"strings"

	"github.com/yixinglu/nebula-importer/pkg/clientmgr"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/csv"
)

type DataFileReader interface {
	Read()
}

func New(file config.File, dataChs []chan clientmgr.Record) DataFileReader {
	switch strings.ToUpper(file.Type) {
	case "CSV":
		return &csv.CSVReader{
			File:    file,
			DataChs: dataChs,
		}
	default:
		log.Fatalf("Wrong file type: %s", file.Type)
		return nil
	}
}
