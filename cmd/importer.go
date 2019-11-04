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

	errCh := make(chan errhandler.ErrData)

	statsMgr := stats.NewStatsMgr()
	defer statsMgr.Close()

	mgr := client.NewNebulaClientMgr(yaml.NebulaClientSettings, errCh, statsMgr.GetStatsChan())
	defer mgr.Close()

	for _, file := range yaml.Files {
		mgr.InitFile(file)
		errWriter := errhandler.New(errCh, statsMgr.GetStatsChan())
		r := reader.New(file, mgr.GetDataChans())
		errWriter.InitFile(file)
		r.Read()
	}
}
