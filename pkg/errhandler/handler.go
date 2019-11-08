package errhandler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type Handler struct {
	errCh      <-chan base.ErrData
	failCh     chan<- stats.Stats
	dataWriter DataWriter
}

func New(file config.File, errCh <-chan base.ErrData, failCh chan<- stats.Stats) (*Handler, error) {
	h := Handler{
		errCh:  errCh,
		failCh: failCh,
	}

	switch strings.ToLower(file.Type) {
	case "csv":
		h.dataWriter = csv.NewErrDataWriter(file.CSV.WithLabel)
	default:
		return nil, fmt.Errorf("Wrong file type: %s", file.Type)
	}

	return &h, nil
}

func mustCreateFile(filePath string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(filePath), 0775); err != nil && !os.IsExist(err) {
		return nil, err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (w *Handler) InitFile(file config.File, concurrency int) error {
	dataFile, err := mustCreateFile(file.Error.FailDataPath)
	if err != nil {
		return err
	}

	logFile, err := mustCreateFile(file.Error.LogPath)
	if err != nil {
		return err
	}

	go func() {
		defer dataFile.Close()
		w.dataWriter.Init(dataFile)

		defer logFile.Close()
		logWriter := bufio.NewWriter(logFile)

		for {
			rawErr := <-w.errCh
			if rawErr.Error == nil {
				concurrency--
				if concurrency == 0 {
					break
				}
			} else {
				w.dataWriter.Write(rawErr.Data)

				logWriter.WriteString(rawErr.Error.Error())
				logWriter.WriteString("\n")

				w.failCh <- stats.NewFailureStats(len(rawErr.Data))
			}
		}

		if err = logWriter.Flush(); err != nil {
			log.Println(err)
		}

		w.dataWriter.Flush()
		if w.dataWriter.Error() != nil {
			log.Println(w.dataWriter.Error())
		}

		w.failCh <- stats.NewFileDoneStats()
	}()

	return nil
}
