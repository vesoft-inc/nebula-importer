package errhandler

import (
	"fmt"
	"os"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type Handler struct {
	statsCh chan<- base.Stats
}

func New(statsCh chan<- base.Stats) *Handler {
	h := Handler{
		statsCh: statsCh,
	}

	return &h
}

func (w *Handler) Init(file *config.File, concurrency int, cleanup bool) (chan base.ErrData, error) {
	var dataWriter DataWriter
	switch strings.ToLower(*file.Type) {
	case "csv":
		dataWriter = csv.NewErrDataWriter(file.CSV)
	default:
		return nil, fmt.Errorf("Wrong file type: %s", *file.Type)
	}

	dataFile := base.MustCreateFile(*file.FailDataPath)
	errCh := make(chan base.ErrData)

	go func() {
		defer func() {
			if err := dataFile.Close(); err != nil {
				logger.Errorf("Fail to close opened error data file: %s", *file.FailDataPath)
			}
			if cleanup {
				if err := os.Remove(*file.FailDataPath); err != nil {
					logger.Errorf("Fail to remove error data file: %s", *file.FailDataPath)
				} else {
					logger.Infof("Error data file has been removed: %s", *file.FailDataPath)
				}
			}
		}()
		defer close(errCh)
		dataWriter.Init(dataFile)

		for {
			rawErr := <-errCh
			if rawErr.Error == nil {
				concurrency--
				if concurrency == 0 {
					break
				}
			} else {
				dataWriter.Write(rawErr.Data)
				logger.Error(rawErr.Error.Error())
				w.statsCh <- base.NewFailureStats(len(rawErr.Data))
			}
		}

		dataWriter.Flush()
		if dataWriter.Error() != nil {
			logger.Error(dataWriter.Error())
		}
		w.statsCh <- base.NewFileDoneStats(*file.Path)
	}()

	return errCh, nil
}
