package main

import (
	"flag"
	"log"
	"strings"
	"time"

	importer "github.com/yixinglu/nebula-importer/importer"
	csv_importer "github.com/yixinglu/nebula-importer/importer/csv"
)

func main() {
	path := flag.String("config", "", "Specify importer configure file path")

	now := time.Now()
	defer func() {
		log.Println("Finish import data, consume time: %.2f", time.Since(now).Seconds())
	}()

	flag.Parse()

	yaml, err := importer.Parse(*path)
	if err != nil {
		log.Fatal(err)
	}

	doneCh := make(chan bool)

	failCh := make(chan bool)
	statsCh := make(chan importer.Stats)
	importer.InitStatsWorker(statsCh, failCh)

	errCh := make(chan importer.ErrData)
	mgr := importer.NebulaClientMgr{
		Config: importer.NebulaClientConfig{
			Address:     yaml.Settings.Connection.Address,
			User:        yaml.Settings.Connection.User,
			Password:    yaml.Settings.Connection.Password,
			Retry:       yaml.Settings.Retry,
			Concurrency: yaml.Settings.Concurrency,
		},
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

			props := make([]importer.Prop, len(file.Schema.Props))
			for i, prop := range file.Schema.Props {
				props[i] = importer.Prop{
					Name: prop.Name,
					Type: prop.Type,
				}
			}
			reader = csv_importer.NewCSVReader(file.Schema.Space, file.Schema.Type, file.Schema.Name, props)
		default:
			log.Fatal("Unsupported file type: %s", file.Type)
		}
		// log.Printf("file struct:\n %#v", file)
		errWriter.SetupErrorHandler()
		reader.InitFileReader(file.Path, stmtChs, doneCh)
	}

	close(statsCh)
}
