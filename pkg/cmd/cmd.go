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

	clientMgr := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	defer clientMgr.Close()

	for _, file := range yaml.Files {
		clientMgr.InitFile(file)

		if handler, err := errhandler.New(file, clientMgr.GetErrChan(), statsMgr.StatsCh); err != nil {
			return err
		} else {
			handler.Init(yaml.NebulaClientSettings.Concurrency)
		}

		r := reader.New(file, clientMgr.GetDataChans())
		if err := r.Read(); err != nil {
			return err
		}

		// Wait to finish handle errors
		<-statsMgr.FileDoneCh
	}

	return nil
}
