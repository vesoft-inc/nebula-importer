package nebula_importer

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
)

type ErrorWriter interface {
	SetupErrorDataHandler()
	SetupErrorLogHandler()
}

type CSVErrWriter struct {
	ErrConf   ErrorConfig
	ErrDataCh <-chan []interface{}
	ErrLogCh  <-chan error
}

func (w *CSVErrWriter) SetupErrorDataHandler() {
	go func() {
		file, err := os.Create(w.ErrConf.ErrorDataPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		for {
			rawErrData := <-w.ErrDataCh
			errData := make([]string, len(rawErrData))
			for i := range rawErrData {
				errData[i] = rawErrData[i].(string)
			}
			writer.Write(errData)
		}
	}()
}

func (w *CSVErrWriter) SetupErrorLogHandler() {
	go func() {
		file, err := os.Create(w.ErrConf.ErrorLogPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		for {
			err := <-w.ErrLogCh
			writer.WriteString(err.Error())
		}
	}()
}
