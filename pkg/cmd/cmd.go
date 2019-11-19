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

func Run(conf string) error {
	yaml, err := config.Parse(conf)
	if err != nil {
		return err
	}

	logger.Init(yaml.LogPath)

	now := time.Now()
	defer func() {
		logger.Log.Printf("Finish import data, consume time: %.2fs", time.Since(now).Seconds())
	}()

	statsMgr := stats.NewStatsMgr(len(yaml.Files))
	defer statsMgr.Close()

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	if err != nil {
		return err
	}
	defer clientMgr.Close()

	errHandler := errhandler.New(statsMgr.StatsCh)

	for _, file := range yaml.Files {
		// TODO: skip files with error
		errCh, err := errHandler.Init(file, yaml.NebulaClientSettings.Concurrency)
		if err != nil {
			return err
		}

		if r, err := reader.New(file, clientMgr.GetRequestChans(), errCh); err != nil {
			return err
		} else {
			go r.Read()
		}
	}

	<-statsMgr.DoneCh

	if statsMgr.NumFailed > 0 {
		return fmt.Errorf("Total %d lines fail to insert to nebula", statsMgr.NumFailed)
	}

	return nil
}
