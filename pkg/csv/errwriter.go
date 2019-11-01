package csv

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/errhandler"
)

type CSVErrWriter struct {
	errConf base.ErrorConfig
	errCh   <-chan base.ErrData
	failCh  chan<- bool
}

func NewCSVErrorWriter(errDataPath, errLogPath string, errCh <-chan base.ErrData, failCh chan<- bool) errhandler.ErrorWriter {
	return &CSVErrWriter{
		errConf: base.ErrorConfig{
			ErrorDataPath: errDataPath,
			ErrorLogPath:  errLogPath,
		},
		errCh:  errCh,
		failCh: failCh,
	}
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
		dataFile := requireFile(w.errConf.ErrorDataPath)
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		logFile := requireFile(w.errConf.ErrorLogPath)
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		for {
			select {
			case rawErr := <-w.errCh:
				if rawErr.Done {
					return
				}

				writeFailedData(dataWriter, rawErr.Data)
				logErrorMessage(logWriter, rawErr.Error)

				w.failCh <- true
			}
		}
	}()

	log.Println("Setup CSV error handler")
}

func writeFailedData(writer *csv.Writer, data [][]interface{}) {
	if len(data) == 0 {
		log.Println("Empty error data")
	}
	record := make([]string, len(data[0]))
	for _, r := range data {
		for i := range r {
			record[i] = r[i].(string)
		}
		writer.Write(record)
	}
}

func logErrorMessage(writer *bufio.Writer, err error) {
	writer.WriteString(err.Error())
	writer.WriteString("\n")
}
