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
		log.Printf("Finish import data, consume time: %.2f", time.Since(now).Seconds())
	}()

	statsMgr := stats.NewStatsMgr()
	defer statsMgr.Close()

	clientMgr := client.NewNebulaClientMgr(yaml.NebulaClientSettings, statsMgr.GetStatsChan())
	defer clientMgr.Close()

	for _, file := range yaml.Files {
		clientMgr.InitFile(file)

		errWriter := errhandler.New(file, yaml.NebulaClientSettings.Concurrency, clientMgr.GetErrChan(), statsMgr.GetStatsChan())
		errWriter.InitFile(file)

		r := reader.New(file, clientMgr.GetDataChans())
		r.Read()

		// Wait to finish handle errors
		<-errWriter.GetFinishChan()
	}
}
