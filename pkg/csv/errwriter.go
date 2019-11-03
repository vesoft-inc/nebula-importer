package csv

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
)

type CSVErrWriter struct {
	ErrConf config.ErrConfig
	ErrCh   <-chan base.ErrData
	FailCh  chan<- bool
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

func (w *CSVErrWriter) SetupErrorHandler() {
	go func() {
		dataFile := requireFile(w.ErrConf.FailDataPath)
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		logFile := requireFile(w.ErrConf.LogPath)
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		for {
			select {
			case rawErr := <-w.ErrCh:
				if rawErr.Done {
					return
				}

				writeFailedData(dataWriter, rawErr.Data)
				logErrorMessage(logWriter, rawErr.Error)

				w.FailCh <- true
			}
		}
	}()

	log.Println("Setup CSV error handler")
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
