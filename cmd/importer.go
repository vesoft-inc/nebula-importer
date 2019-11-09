package main

import (
	"flag"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
)

var configuration = flag.String("config", "", "Specify importer configure file path")

func main() {
	flag.Parse()

	if err := cmd.Run(*configuration); err != nil {
		panic(err)
	}
}
