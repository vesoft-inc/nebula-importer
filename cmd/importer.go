package main

import (
	"flag"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/web"
)

var configuration = flag.String("config", "", "Specify importer configure file path")
var port = flag.Int("port", 5699, "http server port")

func main() {
	flag.Parse()

	if configuration == nil {
		panic("please configure yaml file")
	}

	if port != nil {
		web.Start(*port)
	}

	runner := &cmd.Runner{}
	runner.Run(*configuration)

	if runner.Error() != nil {
		panic(runner.Error())
	}

}
