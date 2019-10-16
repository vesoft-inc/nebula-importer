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
	ErrDataCh <-chan []string
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
			writer.Write(<-w.ErrDataCh)
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
