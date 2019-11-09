package reader

import (
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type DataFileReader interface {
	Read() error
}

func New(file config.File, dataChs []chan base.Data) DataFileReader {
	switch strings.ToUpper(file.Type) {
	case "CSV":
		return &csv.CSVReader{
			File:    file,
			DataChs: dataChs,
		}
	default:
		logger.Log.Fatalf("Wrong file type: %s", file.Type)
		return nil
	}
}
