package cmd

import (
	"fmt"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/client"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errhandler"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/reader"
	"github.com/vesoft-inc/nebula-importer/pkg/stats"
)

type Runner struct {
	err error
}

func (r *Runner) Error() error {
	return r.err
}

func (r *Runner) Run(conf string) {
	now := time.Now()
	defer func() {
		if re := recover(); re != nil {
			r.err = fmt.Errorf("%v", re)
		} else {
			if r.err == nil {
				logger.Infof("Finish import data, consume time: %.2fs", time.Since(now).Seconds())
			}
		}
	}()

	yaml, err := config.Parse(conf)
	if err != nil {
		r.err = err
		return
	}

	logger.Init(*yaml.LogPath)

	statsMgr := stats.NewStatsMgr(len(yaml.Files))
	defer statsMgr.Close()

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	if err != nil {
		r.err = err
		return
	}
	defer clientMgr.Close()

	errHandler := errhandler.New(statsMgr.StatsCh)

	for _, file := range yaml.Files {
		// TODO: skip files with error
		errCh, err := errHandler.Init(file, clientMgr.GetNumConnections())
		if err != nil {
			r.err = err
			return
		}

		if fr, err := reader.New(file, clientMgr.GetRequestChans(), errCh); err != nil {
			r.err = err
			return
		} else {
			go fr.Read()
		}
	}

	<-statsMgr.DoneCh

	if statsMgr.NumFailed > 0 {
		r.err = fmt.Errorf("Total %d lines fail to insert to nebula", statsMgr.NumFailed)
	} else {
		r.err = nil
	}
}
