package csv

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path"

	importer "github.com/yixinglu/nebula-importer/importer"
)

type CSVErrWriter struct {
	errConf importer.ErrorConfig
	errCh   <-chan importer.ErrData
	failCh  chan<- bool
}

func NewCSVErrorWriter(errDataPath, errLogPath string, errCh <-chan importer.ErrData, failCh chan<- bool) importer.ErrorWriter {
	return &CSVErrWriter{
		errConf: importer.ErrorConfig{
			ErrorDataPath: errDataPath,
			ErrorLogPath:  errLogPath,
		},
		errCh:  errCh,
		failCh: failCh,
	}
}

func (w *CSVErrWriter) SetupErrorHandler() {
	go func() {
		if err := os.MkdirAll(path.Dir(w.errConf.ErrorDataPath), 0775); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		dataFile, err := os.Create(w.errConf.ErrorDataPath)
		if err != nil {
			log.Fatal(err)
		}
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		if err := os.MkdirAll(path.Dir(w.errConf.ErrorLogPath), 0775); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		logFile, err := os.Create(w.errConf.ErrorLogPath)
		if err != nil {
			log.Fatal(err)
		}
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		for {
			select {
			case rawErr := <-w.errCh:
				if rawErr.Done {
					return
				}
				// Write failed data
				errData := make([]string, len(rawErr.Data))
				for i := range rawErr.Data {
					errData[i] = rawErr.Data[i].(string)
				}

				dataWriter.Write(errData)

				// Write error message
				logWriter.WriteString(rawErr.Error.Error())
				logWriter.WriteString("\n")

				w.failCh <- true
			}
		}
	}()

	log.Println("Setup CSV error handler")
}
