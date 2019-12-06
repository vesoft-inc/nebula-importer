package csv

import (
	"encoding/csv"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type ErrWriter struct {
	writer    *csv.Writer
	csvConfig *config.CSVConfig
}

func NewErrDataWriter(config *config.CSVConfig) *ErrWriter {
	return &ErrWriter{
		csvConfig: config,
	}
}

func (w *ErrWriter) Error() error {
	return w.writer.Error()
}

func (w *ErrWriter) Init(f *os.File) {
	w.writer = csv.NewWriter(f)
}

func (w *ErrWriter) Write(data []base.Data) {
	if len(data) == 0 {
		logger.Info("Empty error data")
	}
	for _, d := range data {
		if *w.csvConfig.WithLabel {
			var record []string
			switch d.Type {
			case base.INSERT:
				record = append(record, "+")
			case base.DELETE:
				record = append(record, "-")
			default:
				logger.Fatalf("Error data type: %s", d.Type)
			}
			record = append(record, d.Record...)
			w.writer.Write(record)
		} else {
			w.writer.Write(d.Record)
		}
	}
}

func (w *ErrWriter) Flush() {
	w.writer.Flush()
}
