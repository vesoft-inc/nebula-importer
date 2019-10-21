package csv

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path"
	"time"

	importer "github.com/yixinglu/nebula-importer/importer"
)

type CSVErrWriter struct {
	ErrConf importer.ErrorConfig
	ErrCh   <-chan importer.ErrData
}

func NewCSVErrorWriter(errDataPath, errLogPath string, errCh <-chan importer.ErrData) importer.ErrorWriter {
	return &CSVErrWriter{
		ErrConf: importer.ErrorConfig{
			ErrorDataPath: errDataPath,
			ErrorLogPath:  errLogPath,
		},
		ErrCh: errCh,
	}
}

func (w *CSVErrWriter) SetupErrorHandler() {
	go func() {
		if err := os.MkdirAll(path.Dir(w.ErrConf.ErrorDataPath), 0775); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		dataFile, err := os.Create(w.ErrConf.ErrorDataPath)
		if err != nil {
			log.Fatal(err)
		}
		defer dataFile.Close()

		dataWriter := csv.NewWriter(dataFile)

		if err := os.MkdirAll(path.Dir(w.ErrConf.ErrorLogPath), 0775); err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		logFile, err := os.Create(w.ErrConf.ErrorLogPath)
		if err != nil {
			log.Fatal(err)
		}
		defer logFile.Close()

		logWriter := bufio.NewWriter(logFile)

		ticker := time.NewTicker(30 * time.Second)

		var numFailed uint64 = 0
		for {
			select {
			case <-ticker.C:
				log.Printf("Failed queries: %d", numFailed)
			case rawErr := <-w.ErrCh:
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

				numFailed++
			}
		}
	}()

	log.Println("Setup CSV error handler")
}
