package reader

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
)

type DataFileReader interface {
	Read() error
}

func New(file config.File, dataChs []chan base.Data) (DataFileReader, error) {
	switch strings.ToUpper(file.Type) {
	case "CSV":
		r := csv.CSVReader{
			File:    file,
			DataChs: dataChs,
		}
		return &r, nil
	default:
		return nil, fmt.Errorf("Wrong file type: %s", file.Type)
	}
}
