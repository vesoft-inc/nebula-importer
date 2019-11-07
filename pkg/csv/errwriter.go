package csv

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type CSVErrWriter struct {
	ErrCh  <-chan base.ErrData
	FailCh chan<- stats.Stats
	file   config.File
}

func mustCreateFile(filePath string) *os.File {
	if err := os.MkdirAll(path.Dir(filePath), 0775); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func (w *CSVErrWriter) InitFile(file config.File, concurrency int) {
	w.file = file
	go func() {
		dataFile := mustCreateFile(file.Error.FailDataPath)
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		logFile := mustCreateFile(file.Error.LogPath)
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		for {
			rawErr := <-w.ErrCh
			if rawErr.Error == nil {
				concurrency--
				if concurrency == 0 {
					break
				} else {
					continue
				}
			}

			w.writeFailedData(dataWriter, rawErr.Data)
			w.logErrorMessage(logWriter, rawErr.Error)

			w.FailCh <- stats.NewFailureStats(len(rawErr.Data))
		}

		logWriter.Flush()

		dataWriter.Flush()
		if err := dataWriter.Error(); err != nil {
			log.Fatal(err)
		}

		w.FailCh <- stats.NewFileDoneStats()
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
