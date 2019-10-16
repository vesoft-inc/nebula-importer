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

	stmtCh := make(chan importer.Stmt)
	errLogCh := make(chan error)
	errDataCh := make(chan []interface{})
	clientConf := importer.NebulaClientConfig{
		Address:     yaml.Settings.Connection.Address,
		User:        yaml.Settings.Connection.User,
		Password:    yaml.Settings.Connection.Password,
		Retry:       yaml.Settings.Retry,
		Concurrency: yaml.Settings.Concurrency,
	}
	importer.InitNebulaClientPool(clientConf, stmtCh, errLogCh, errDataCh)

	for _, file := range yaml.Files {
		// Setup error handler
		var errorWriter importer.ErrorWriter
		csvErrWriter := importer.CSVErrWriter{
			ErrConf: importer.ErrorConfig{
				ErrorDataPath: file.Error.FailDataPath,
				ErrorLogPath:  file.Error.LogPath,
			},
			ErrDataCh: errDataCh,
			ErrLogCh:  errLogCh,
		}

		errorWriter = &csvErrWriter
		errorWriter.SetupErrorDataHandler()
		errorWriter.SetupErrorLogHandler()

		// Setup reader
		var reader importer.DataFileReader
		if strings.ToLower(file.Type) == "csv" {
			csvReader := importer.CSVReader{
				Schema: importer.Schema{
					Type: file.Schema.Type,
					Name: file.Schema.Name,
				},
			}
			for _, prop := range file.Schema.Props {
				csvReader.Schema.Props = append(csvReader.Schema.Props, importer.Prop{
					Name: prop.Name,
					Type: prop.Type,
				})
			}
			reader = &csvReader
		}
		reader.NewFileReader(file.Path, stmtCh)
	}
}
