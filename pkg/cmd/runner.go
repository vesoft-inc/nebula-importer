package cmd

import (
	"errors"
	"fmt"
	"sync"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/client"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errhandler"
	importerError "github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/reader"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type Runner struct {
	errs      []error
	Readers   []*reader.FileReader
	stataMgr  *stats.StatsMgr
	NumFailed int64
}

func (r *Runner) Error() error {
	if len(r.errs) == 0 {
		return nil
	}

	// TODO(yee): Only return first error
	return r.errs[0]
}

func (r *Runner) Errors() []error {
	return r.errs
}

func (r *Runner) Run(yaml *config.YAMLConfig) {
	defer func() {
		if re := recover(); re != nil {
			r.errs = append(r.errs, importerError.Wrap(importerError.UnknownError, fmt.Errorf("%v", re)))
		}
	}()

	runnerLogger := logger.NewRunnerLogger(*yaml.LogPath)

	statsMgr := stats.NewStatsMgr(yaml.Files, runnerLogger)
	defer statsMgr.Close()

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh, runnerLogger)
	if err != nil {
		r.errs = append(r.errs, importerError.Wrap(importerError.NebulaError, err))
		return
	}
	defer clientMgr.Close()

	errHandler := errhandler.New(statsMgr.StatsCh)

	freaders := make([]*reader.FileReader, len(yaml.Files))

	var wgReaders sync.WaitGroup
	for i, file := range yaml.Files {
		errCh, err := errHandler.Init(file, clientMgr.GetNumConnections(), *yaml.RemoveTempFiles, runnerLogger)
		if err != nil {
			r.errs = append(r.errs, importerError.Wrap(importerError.ConfigError, err))
			statsMgr.StatsCh <- base.NewFileDoneStats(*file.Path)
			continue
		}

		if fr, err := reader.New(i, file, *yaml.RemoveTempFiles, clientMgr.GetRequestChans(), errCh, runnerLogger); err != nil {
			r.errs = append(r.errs, importerError.Wrap(importerError.ConfigError, err))
			statsMgr.StatsCh <- base.NewFileDoneStats(*file.Path)
			continue
		} else {
			runnerLogger.Infof("Start to read %s", *file.Path)
			wgReaders.Add(1)
			go func(fr *reader.FileReader, filename string) {
				defer func() {
					runnerLogger.Infof("Finish to read %s", filename)
					wgReaders.Done()
				}()
				numReadFailed, err := fr.Read()
				statsMgr.Stats.NumReadFailed += numReadFailed
				if err != nil {
					r.errs = append(r.errs, err)
					statsMgr.StatsCh <- base.NewFileDoneStats(filename)
				}
			}(fr, *file.Path)
			freaders[i] = fr
		}
	}

	r.Readers = freaders
	r.stataMgr = statsMgr

	runnerLogger.Infof("Waiting for stats manager done")
	<-statsMgr.DoneCh
	runnerLogger.Infof("Waiting for all readers exit")
	for _, r := range freaders {
		if r != nil {
			r.Stop()
		}
	}
	// fix issues/219
	// The number of times `statsMgr.StatsCh <- base.NewFileDoneStats(filename)` has reached the number of readers,
	// then <-statsMgr.DoneCh return, but not all readers have exited.
	// So, it's need to wait for it exit.
	wgReaders.Wait()
	runnerLogger.Infof("All readers exited")

	r.stataMgr.CountFileBytes(r.Readers)
	r.Readers = nil
	r.NumFailed = statsMgr.Stats.NumFailed

	if statsMgr.Stats.NumFailed > 0 {
		r.errs = append(r.errs, importerError.Wrap(importerError.NotCompleteError,
			fmt.Errorf("Total %d lines fail to insert into nebula graph database", statsMgr.Stats.NumFailed)))
	}
}

func (r *Runner) QueryStats() (*stats.Stats, error) {
	if r.stataMgr != nil {
		if r.Readers != nil {
			err := r.stataMgr.CountFileBytes(r.Readers)
			if err != nil {
				return nil, importerError.Wrap(importerError.NotCompleteError, err)
			}
		}
		if r.stataMgr.Done == true {
			return &r.stataMgr.Stats, nil
		}
		r.stataMgr.StatsCh <- base.NewOutputStats()
		select {
		case stats, ok := <-r.stataMgr.OutputStatsCh:
			if !ok {
				return nil, importerError.Wrap(importerError.UnknownError, errors.New("output stats to chanel fail"))
			}
			return &stats, nil
		}
	} else {
		return nil, importerError.Wrap(importerError.NotCompleteError, errors.New("stataMgr not init complete"))
	}
}
