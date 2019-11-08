package errhandler

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type Handler struct {
	file       config.File
	errCh      <-chan base.ErrData
	failCh     chan<- stats.Stats
	dataWriter DataWriter
}

func New(file config.File, errCh <-chan base.ErrData, failCh chan<- stats.Stats) (*Handler, error) {
	h := Handler{
		file:   file,
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

func (w *Handler) Init(concurrency int) {
	dataFile := base.MustCreateFile(w.file.FailDataPath)

	go func() {
		defer dataFile.Close()
		w.dataWriter.Init(dataFile)

		for {
			rawErr := <-w.errCh
			if rawErr.Error == nil {
				concurrency--
				if concurrency == 0 {
					break
				}
			} else {
				w.dataWriter.Write(rawErr.Data)

				logger.Log.Println(rawErr.Error.Error())

				w.failCh <- stats.NewFailureStats(len(rawErr.Data))
			}
		}

		w.dataWriter.Flush()
		if w.dataWriter.Error() != nil {
			logger.Log.Println(w.dataWriter.Error())
		}

		w.failCh <- stats.NewFileDoneStats()
	}()
}
