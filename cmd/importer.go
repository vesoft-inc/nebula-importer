package main

import (
	"flag"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/web"
)

var configuration = flag.String("config", "", "Specify importer configure file path")
var port = flag.Int("port", -1, "HTTP server port")
var callback = flag.String("callback", "", "HTTP server callback address")

func main() {
	flag.Parse()

	if port != nil && *port > 0 && callback != nil && *callback != "" {
		// Start http server
		svr := &web.WebServer{
			Port:     *port,
			Callback: *callback,
		}

		svr.Start()
	} else {
		if configuration == nil {
			panic("please configure yaml file")
		}

		conf, err := config.Parse(*configuration)
		if err != nil {
			e := err.(errors.ImporterError)
			os.Exit(e.ErrCode)
		}

		runner := &cmd.Runner{}
		runner.Run(conf)

		if runner.Error() != nil {
			e := runner.Error().(errors.ImporterError)
			os.Exit(e.ErrCode)
		}
	}
}
