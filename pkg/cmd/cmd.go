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

	clientMgr, err := client.NewNebulaClientMgr(yaml.NebulaClientSettings)
	if err != nil {
		return err
	}
	defer clientMgr.Close()

	errHandler := errhandler.New(statsMgr.StatsCh)

	for _, file := range yaml.Files {
		// TODO: skip files with error
		if err := errHandler.Init(file, yaml.NebulaClientSettings.Concurrency); err != nil {
			return err
		}

		if r, err := reader.New(file, clientMgr.GetRequestChans(), statsMgr.StatsCh, errHandler.ErrCh); err != nil {
			return err
		} else {
			if err := r.Read(); err != nil {
				return err
			}
		}
	}

	return nil
}
