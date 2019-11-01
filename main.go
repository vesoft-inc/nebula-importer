package main

import (
	"flag"
	"log"
	"strings"
	"time"

	importer "github.com/yixinglu/nebula-importer/importer"
	csv_importer "github.com/yixinglu/nebula-importer/importer/csv"
)

var config = flag.String("config", "", "Specify importer configure file path")

func main() {
	flag.Parse()

	yaml, err := importer.Parse(*config)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	defer func() {
		log.Printf("Finish import data, consume time: %.2f", time.Since(now).Seconds())
	}()

	doneCh := make(chan bool)

	failCh := make(chan bool)
	statsCh := make(chan importer.Stats)
	importer.InitStatsWorker(statsCh, failCh)

	errCh := make(chan importer.ErrData)
	mgr := importer.NebulaClientMgr{
		Config:  yaml.NebulaClientSettings,
		ErrCh:   errCh,
		StatsCh: statsCh,
		DoneCh:  doneCh,
	}
	stmtChs := mgr.InitNebulaClientPool()

	for _, file := range yaml.Files {
		var errWriter importer.ErrorWriter
		var reader importer.DataFileReader
		switch strings.ToLower(file.Type) {
		case "csv":
			errWriter = csv_importer.NewCSVErrorWriter(file.Error.FailDataPath, file.Error.LogPath, errCh, failCh)

			descs := make([]importer.Desc, len(file.Schema.Descs))
			for i, desc := range file.Schema.Descs {
				descs[i] = importer.Desc{Name: desc.Name}
				props := make([]importer.Prop, len(desc.Props))
				for j, prop := range desc.Props {
					props[j] = importer.Prop{
						Name: prop.Name,
						Type: prop.Type,
					}
				}
			}
			reader = csv_importer.NewCSVReader(file.Schema.Space, file.Schema.Type, descs)
		default:
			log.Fatal("Unsupported file type: %s", file.Type)
		}
		// log.Printf("file struct:\n %#v", file)
		errWriter.SetupErrorHandler()
		reader.InitFileReader(stmtChs, doneCh)
	}

	close(statsCh)
}
