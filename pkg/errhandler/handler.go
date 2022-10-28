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

func (w *Handler) Init(file *config.File, concurrency int, cleanup bool, runnerLogger *logger.RunnerLogger) (chan base.ErrData, error) {
	var dataWriter DataWriter
	switch strings.ToLower(*file.Type) {
	case "csv":
		dataWriter = csv.NewErrDataWriter(file.CSV, runnerLogger)
	default:
		return nil, fmt.Errorf("Wrong file type: %s", *file.Type)
	}

	dataFile := base.MustCreateFile(*file.FailDataPath)
	errCh := make(chan base.ErrData)

	go func() {
		defer func() {
			if err := dataFile.Close(); err != nil {
				logger.Log.Errorf("Fail to close opened error data file: %s", *file.FailDataPath)
			}
			if cleanup {
				if err := os.Remove(*file.FailDataPath); err != nil {
					logger.Log.Errorf("Fail to remove error data file: %s", *file.FailDataPath)
				} else {
					logger.Log.Infof("Error data file has been removed: %s", *file.FailDataPath)
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
				logger.Log.Error(rawErr.Error.Error())
				var importedBytes int64
				for _, d := range rawErr.Data {
					importedBytes += int64(d.Bytes)
				}
				w.statsCh <- base.NewFailureStats(len(rawErr.Data), importedBytes)
			}
		}

		dataWriter.Flush()
		if dataWriter.Error() != nil {
			logger.Log.Error(dataWriter.Error())
		}
		w.statsCh <- base.NewFileDoneStats(*file.Path)
	}()

	return errCh, nil
}
