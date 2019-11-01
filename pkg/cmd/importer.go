package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/clientmgr"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/csv"
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

	doneCh := make(chan bool)

	failCh := make(chan bool)
	statsCh := make(chan stats.Stats)
	stats.InitStatsWorker(statsCh, failCh)

	errCh := make(chan base.ErrData)
	mgr := clientmgr.NebulaClientMgr{
		Config:  yaml.NebulaClientSettings,
		ErrCh:   errCh,
		StatsCh: statsCh,
		DoneCh:  doneCh,
	}
	stmtChs := mgr.InitNebulaClientPool()

	for _, file := range yaml.Files {
		var errWriter errhandler.ErrorWriter
		var reader reader.DataFileReader
		switch strings.ToLower(file.Type) {
		case "csv":
			errWriter = csv.NewCSVErrorWriter(file.Error, errCh, failCh)
			reader = csv.NewCSVReader(file)
		default:
			log.Fatal("Unsupported file type: %s", file.Type)
		}
		errWriter.SetupErrorHandler()
		reader.InitFileReader(stmtChs, doneCh)
	}

	close(statsCh)
}
