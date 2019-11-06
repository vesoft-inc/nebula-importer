package cmd

import (
	"fmt"
	"time"

	"github.com/yixinglu/nebula-importer/pkg/client"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/errhandler"
	"github.com/yixinglu/nebula-importer/pkg/reader"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

func Run(conf string) error {
	yaml, err := config.Parse(conf)
	if err != nil {
		return err
	}

	now := time.Now()
	defer func() {
		fmt.Printf("\nFinish import data, consume time: %.2fs\n", time.Since(now).Seconds())
	}()

	statsMgr := stats.NewStatsMgr()
	defer statsMgr.Close()

	clientMgr := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.StatsCh)
	defer clientMgr.Close()

	for _, file := range yaml.Files {
		clientMgr.InitFile(file)

		errWriter := errhandler.New(file, clientMgr.GetErrChan(), statsMgr.StatsCh)
		errWriter.InitFile(file, yaml.NebulaClientSettings.Concurrency)

		r := reader.New(file, clientMgr.GetDataChans())
		r.Read()

		// Wait to finish handle errors
		<-statsMgr.FileDoneCh
	}

	return nil
}
