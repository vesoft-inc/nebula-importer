package main

import (
	"flag"
	"log"
	"time"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/clientmgr"
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

	failCh := make(chan bool)
	statsCh := make(chan stats.Stats)
	stats.InitStatsWorker(statsCh, failCh)

	mgr := clientmgr.NewNebulaClientMgr(yaml.NebulaClientSettings)
	defer mgr.Close()

	errCh := make(chan base.ErrData)
	for _, file := range yaml.Files {
		errWriter := errhandler.New(file, errCh, failCh)
		r := reader.New(file, mgr.GetDataChans())
		errWriter.SetupErrorHandler()
		r.Read()
	}

	close(statsCh)
}
