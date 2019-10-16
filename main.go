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

	log.Printf("%v", yaml)

	stmtCh := make(chan string)
	errCh := make(chan error)
	clientConf := importer.NebulaClientConfig{
		Address:     yaml.Settings.Connection.Address,
		User:        yaml.Settings.Connection.User,
		Password:    yaml.Settings.Connection.Password,
		Retry:       yaml.Settings.Retry,
		Concurrency: yaml.Settings.Concurrency,
	}
	importer.InitNebulaClientPool(clientConf, stmtCh, errCh)

	for _, file := range yaml.Files {
		if strings.ToLower(file.Type) == "csv" {
			reader := importer.CSVReader{
				Schema: importer.Schema{
					Type: file.Schema.Type,
					Name: file.Schema.Name,
				},
			}
			for _, prop := range file.Schema.Props {
				reader.Schema.Props = append(reader.Schema.Props, importer.Prop{
					Name: prop.Name,
					Type: prop.Type,
				})
			}
			reader.NewFileReader(file.Path, stmtCh)
		}
	}
}
