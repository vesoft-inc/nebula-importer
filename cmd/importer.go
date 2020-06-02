package main

import (
	"flag"
	"log"
	"os"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/web"
)

var configuration = flag.String("config", "", "Specify importer configure file path")
var port = flag.Int("port", -1, "HTTP server port")
var callback = flag.String("callback", "", "HTTP server callback address")
var cleanup = flag.Bool("remove_temp_files", false, "Whether to cleanup all generated temporary log and data files")

func main() {
	flag.Parse()

	if port != nil && *port > 0 && callback != nil && *callback != "" {
		// Start http server
		svr := &web.WebServer{
			Port:     *port,
			Callback: *callback,
			Cleanup:  false,
		}

		svr.Start()
	} else {
		if configuration == nil {
			panic("please configure yaml file")
		}

		conf, err := config.Parse(*configuration)
		if err != nil {
			e := err.(errors.ImporterError)
			log.Println(e.ErrMsg.Error())
			os.Exit(e.ErrCode)
		}

		runner := &cmd.Runner{
			Cleanup: *cleanup,
		}
		runner.Run(conf)

		if runner.Error() != nil {
			e := runner.Error().(errors.ImporterError)
			log.Println(e.ErrMsg.Error())
			os.Exit(e.ErrCode)
		}
	}
}
