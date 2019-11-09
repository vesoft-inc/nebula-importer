package cmd

import (
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

	statsMgr := stats.NewStatsMgr()
	defer statsMgr.Close()

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	if err != nil {
		return err
	}
	defer clientMgr.Close()

	for _, file := range yaml.Files {
		if err := clientMgr.InitFile(file); err != nil {
			return err
		}

		if handler, err := errhandler.New(file, clientMgr.GetErrChan(), statsMgr.StatsCh); err != nil {
			return err
		} else {
			handler.Init(yaml.NebulaClientSettings.Concurrency)
		}

		if r, err := reader.New(file, clientMgr.GetDataChans()); err != nil {
			return err
		} else {
			if err := r.Read(); err != nil {
				return err
			}
		}

		// Wait to finish handle errors
		<-statsMgr.FileDoneCh
	}

	return nil
}
