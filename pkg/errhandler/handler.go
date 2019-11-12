package errhandler

import (
	"fmt"
	"strings"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/csv"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type Handler struct {
	ErrCh   chan base.ErrData
	statsCh chan<- base.Stats
	logCh   chan error
}

func New(statsCh chan<- base.Stats) *Handler {
	h := Handler{
		ErrCh:   make(chan base.ErrData),
		statsCh: statsCh,
		logCh:   make(chan error),
	}

	go h.startErrorLogWorker()

	return &h
}

func (w *Handler) startErrorLogWorker() {
	for {
		err := <-w.logCh
		logger.Log.Println(err.Error())
	}
}

func (w *Handler) Init(file config.File, concurrency int) error {
	var dataWriter DataWriter
	switch strings.ToLower(file.Type) {
	case "csv":
		dataWriter = csv.NewErrDataWriter(file.CSV)
	default:
		return fmt.Errorf("Wrong file type: %s", file.Type)
	}

	dataFile := base.MustCreateFile(file.FailDataPath)

	go func() {
		defer dataFile.Close()
		dataWriter.Init(dataFile)

		for {
			rawErr := <-w.ErrCh
			if rawErr.Error == nil {
				concurrency--
				if concurrency == 0 {
					break
				}
			} else {
				dataWriter.Write(rawErr.Data)
				w.logCh <- rawErr.Error
				w.statsCh <- base.NewFailureStats(len(rawErr.Data))
			}
		}

		dataWriter.Flush()
		if dataWriter.Error() != nil {
			w.logCh <- dataWriter.Error()
		}
	}()

	return nil
}
