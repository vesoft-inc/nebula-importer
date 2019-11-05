package main

import (
	"flag"
	"log"
	"time"

	"github.com/yixinglu/nebula-importer/pkg/client"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/errhandler"
	"github.com/yixinglu/nebula-importer/pkg/reader"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

var configuration = flag.String("config", "", "Specify importer configure file path")

func main() {
	flag.Parse()

	yaml, err := config.Parse(*configuration)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	defer func() {
		log.Printf("\nFinish import data, consume time: %.2fs", time.Since(now).Seconds())
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
}
