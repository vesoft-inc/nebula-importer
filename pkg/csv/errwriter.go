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

			writeFailedData(dataWriter, rawErr.Data.Record)
			logErrorMessage(logWriter, rawErr.Error)

			w.FailCh <- stats.NewFailureStats()
		}
	}()
}

func writeFailedData(writer *csv.Writer, data base.Record) {
	if len(data) == 0 {
		log.Println("Empty error data")
	}
	writer.Write(data)
}

func logErrorMessage(writer *bufio.Writer, err error) {
	writer.WriteString(err.Error())
	writer.WriteString("\n")
}
