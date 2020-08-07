package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/web"
)

var configuration = flag.String("config", "", "Specify importer configure file path")
var port = flag.Int("port", -1, "HTTP server port")
var callback = flag.String("callback", "", "HTTP server callback address")

func main() {
	errCode := 0
	defer func() {
		// Just for filebeat log fetcher to differentiate following logs from others
		time.Sleep(1 * time.Second)
		log.Println("--- END OF NEBULA IMPORTER ---")
		os.Exit(errCode)
	}()

	log.Println("--- START OF NEBULA IMPORTER ---")

	flag.Parse()

	if port != nil && *port > 0 && callback != nil && *callback != "" {
		// Start http server
		svr := &web.WebServer{
			Port:     *port,
			Callback: *callback,
		}

		if err := svr.Start(); err != nil {
			panic(err)
		}
	} else {
		if configuration == nil {
			panic("please configure yaml file")
		}

		conf, err := config.Parse(*configuration)
		if err != nil {
			e := err.(errors.ImporterError)
			log.Println(e.ErrMsg.Error())
			errCode = e.ErrCode
			return
		}

		runner := &cmd.Runner{}

		{
			now := time.Now()
			defer func() {
				time.Sleep(500 * time.Millisecond)
				if runner.Error() != nil {
					e := runner.Error().(errors.ImporterError)
					errCode = e.ErrCode
					log.Println(e.ErrMsg.Error())
				} else {
					log.Printf("Finish import data, consume time: %.2fs", time.Since(now).Seconds())
				}
			}()

			runner.Run(conf)
		}
	}
}
