package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/client"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errhandler"
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

	var msg []string
	for _, e := range r.errs {
		msg = append(msg, e.Error())
	}
	return errors.New(strings.Join(msg, "\n"))
}

func (r *Runner) Run(yaml *config.YAMLConfig) {
	now := time.Now()
	defer func() {
		if re := recover(); re != nil {
			r.errs = append(r.errs, fmt.Errorf("%v", re))
		} else {
			if len(r.errs) == 0 {
				logger.Infof("Finish import data, consume time: %.2fs", time.Since(now).Seconds())
			}
		}
	}()

	logger.Init(*yaml.LogPath)

	statsMgr := stats.NewStatsMgr(len(yaml.Files))
	defer statsMgr.Close()

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	if err != nil {
		r.errs = append(r.errs, err)
		return
	}
	defer clientMgr.Close()

	errHandler := errhandler.New(statsMgr.StatsCh)

	freaders := make([]*reader.FileReader, len(yaml.Files))

	for i, file := range yaml.Files {
		errCh, err := errHandler.Init(file, clientMgr.GetNumConnections())
		if err != nil {
			r.errs = append(r.errs, err)
			statsMgr.StatsCh <- base.NewFileDoneStats(*file.Path)
			continue
		}

		if fr, err := reader.New(i, file, clientMgr.GetRequestChans(), errCh); err != nil {
			r.errs = append(r.errs, err)
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
		r.errs = append(r.errs, fmt.Errorf("Total %d lines fail to insert to nebula", statsMgr.NumFailed))
	}
}
