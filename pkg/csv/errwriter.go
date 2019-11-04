package csv

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

type CSVErrWriter struct {
	ErrCh       <-chan base.ErrData
	FailCh      chan<- stats.Stats
	Concurrency int
	FinishCh    chan bool
	file        config.File
}

func requireFile(filePath string) *os.File {
	if err := os.MkdirAll(path.Dir(filePath), 0775); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func (w *CSVErrWriter) GetFinishChan() <-chan bool {
	return w.FinishCh
}

func (w *CSVErrWriter) InitFile(file config.File) {
	w.file = file
	go func() {
		dataFile := requireFile(file.Error.FailDataPath)
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		logFile := requireFile(file.Error.LogPath)
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		for {
			rawErr := <-w.ErrCh
			if rawErr.Error == nil {
				w.Concurrency--
				if w.Concurrency == 0 {
					w.FinishCh <- true
					break
				} else {
					continue
				}
			}

			w.writeFailedData(dataWriter, rawErr.Data)
			w.logErrorMessage(logWriter, rawErr.Error)

			w.FailCh <- stats.NewFailureStats(len(rawErr.Data))
		}
	}()
}

func (w *CSVErrWriter) writeFailedData(writer *csv.Writer, data []base.Data) {
	if len(data) == 0 {
		log.Println("Empty error data")
	}
	for _, d := range data {
		if w.file.CSV.WithLabel {
			var record []string
			switch d.Type {
			case base.INSERT:
				record = append(record, "+")
			case base.DELETE:
				record = append(record, "-")
			default:
				log.Fatalf("Error data type: %s", d.Type)
			}
			record = append(record, d.Record...)
			writer.Write(record)
		} else {
			writer.Write(d.Record)
		}
	}
}

func (w *CSVErrWriter) logErrorMessage(writer *bufio.Writer, err error) {
	writer.WriteString(err.Error())
	writer.WriteString("\n")
}
