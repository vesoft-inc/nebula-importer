package csv

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type ErrWriter struct {
	writer       *csv.Writer
	csvConfig    *config.CSVConfig
	runnerLogger *logger.RunnerLogger
}

func NewErrDataWriter(config *config.CSVConfig, runnerLogger *logger.RunnerLogger) *ErrWriter {
	return &ErrWriter{
		csvConfig:    config,
		runnerLogger: runnerLogger,
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
		logger.Log.Info("Empty error data")
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
				logger.Log.Errorf("Error data type: %s, data: %s", d.Type, strings.Join(d.Record, ","))
				continue
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
