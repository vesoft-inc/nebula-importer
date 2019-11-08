package csv

import (
	"encoding/csv"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type ErrWriter struct {
	writer    *csv.Writer
	withLabel bool
}

func NewErrDataWriter(withLabel bool) *ErrWriter {
	return &ErrWriter{
		withLabel: withLabel,
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
		logger.Log.Println("Empty error data")
	}
	for _, d := range data {
		if w.withLabel {
			var record []string
			switch d.Type {
			case base.INSERT:
				record = append(record, "+")
			case base.DELETE:
				record = append(record, "-")
			default:
				logger.Log.Fatalf("Error data type: %s", d.Type)
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
