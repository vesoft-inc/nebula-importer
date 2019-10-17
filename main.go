package main

import (
	"flag"
	"log"
	"strings"

	importer "github.com/yixinglu/nebula-importer/importer"
)

func main() {
	path := flag.String("config", "", "Specify importer configure file path")
	flag.Parse()

	yaml, err := importer.Parse(*path)
	if err != nil {
		log.Fatal(err)
	}

	errLogCh := make(chan error)
	errDataCh := make(chan []interface{})
	clientConf := importer.NebulaClientConfig{
		Address:     yaml.Settings.Connection.Address,
		User:        yaml.Settings.Connection.User,
		Password:    yaml.Settings.Connection.Password,
		Retry:       yaml.Settings.Retry,
		Concurrency: yaml.Settings.Concurrency,
	}
	stmtChs := importer.InitNebulaClientPool(clientConf, errLogCh, errDataCh)

	for _, file := range yaml.Files {
		var errWriter importer.ErrorWriter
		var reader importer.DataFileReader
		switch strings.ToLower(file.Type) {
		case "csv":
			// Setup error handler
			errWriter = &importer.CSVErrWriter{
				ErrConf: importer.ErrorConfig{
					ErrorDataPath: file.Error.FailDataPath,
					ErrorLogPath:  file.Error.LogPath,
				},
				ErrDataCh: errDataCh,
				ErrLogCh:  errLogCh,
			}

			// Setup reader
			csvReader := importer.CSVReader{
				Schema: importer.Schema{
					Space: file.Schema.Space,
					Type:  file.Schema.Type,
					Name:  file.Schema.Name,
				},
			}
			for _, prop := range file.Schema.Props {
				csvReader.Schema.Props = append(csvReader.Schema.Props, importer.Prop{
					Name: prop.Name,
					Type: prop.Type,
				})
			}
			reader = &csvReader
		default:
			log.Fatal("Unsupported file type: %s", file.Type)
		}
		// log.Printf("file struct:\n %#v", file)
		setUpErrorWriter(errWriter)
		reader.NewFileReader(file.Path, stmtChs)
	}
}

func setUpErrorWriter(w importer.ErrorWriter) {
	w.SetupErrorDataHandler()
	w.SetupErrorLogHandler()
}
