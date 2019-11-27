package main

import (
	"flag"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
)

var configuration = flag.String("config", "", "Specify importer configure file path")

func main() {
	flag.Parse()

	runner := &cmd.Runner{}
	runner.Run(*configuration)

	if runner.Error() != nil {
		panic(runner.Error())
	}
}
