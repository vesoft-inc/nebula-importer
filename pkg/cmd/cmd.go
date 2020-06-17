package cmd

import (
	"fmt"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/client"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errhandler"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/reader"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type Runner struct {
	errs      []error
	Readers   []*reader.FileReader
	NumFailed int64
}

func (r *Runner) Error() error {
	if len(r.errs) == 0 {
		return nil
	}

	// TODO(yee): Only return first error
	return r.errs[0]
}

func (r *Runner) Run(yaml *config.YAMLConfig) {
	now := time.Now()
	defer func() {
		if re := recover(); re != nil {
			r.errs = append(r.errs, errors.Wrap(errors.UnknownError, fmt.Errorf("%v", re)))
		} else {
			if len(r.errs) == 0 {
				logger.Infof("Finish import data, consume time: %.2fs", time.Since(now).Seconds())
			}
		}
	}()

	if !*yaml.RemoveTempFiles {
		logger.Init(*yaml.LogPath)
	}

	statsMgr := stats.NewStatsMgr(len(yaml.Files))
	defer statsMgr.Close()

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	if err != nil {
		r.errs = append(r.errs, errors.Wrap(errors.NebulaError, err))
		return
	}
	defer clientMgr.Close()

	errHandler := errhandler.New(statsMgr.StatsCh)

	freaders := make([]*reader.FileReader, len(yaml.Files))

	for i, file := range yaml.Files {
		errCh, err := errHandler.Init(file, clientMgr.GetNumConnections(), *yaml.RemoveTempFiles)
		if err != nil {
			r.errs = append(r.errs, errors.Wrap(errors.ConfigError, err))
			statsMgr.StatsCh <- base.NewFileDoneStats(*file.Path)
			continue
		}

		if fr, err := reader.New(i, file, *yaml.RemoveTempFiles, clientMgr.GetRequestChans(), errCh); err != nil {
			r.errs = append(r.errs, errors.Wrap(errors.ConfigError, err))
			statsMgr.StatsCh <- base.NewFileDoneStats(*file.Path)
			continue
		} else {
			go func(fr *reader.FileReader, filename string) {
				if err := fr.Read(); err != nil {
					r.errs = append(r.errs, err)
					statsMgr.StatsCh <- base.NewFileDoneStats(filename)
				}
			}(fr, *file.Path)
			freaders[i] = fr
		}
	}

	r.Readers = freaders

	<-statsMgr.DoneCh

	r.Readers = nil
	r.NumFailed = statsMgr.NumFailed

	if statsMgr.NumFailed > 0 {
		r.errs = append(r.errs, errors.Wrap(errors.NotCompleteError,
			fmt.Errorf("Total %d lines fail to insert into nebula graph database", statsMgr.NumFailed)))
	}
}
