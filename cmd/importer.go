package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vesoft-inc/nebula-importer/v3/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/config"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/version"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/web"
)

var configuration = flag.String("config", "", "Specify importer configure file path")
var echoVersion = flag.Bool("version", false, "echo build version")
var port = flag.Int("port", -1, "HTTP server port")
var callback = flag.String("callback", "", "HTTP server callback address")

func main() {
	errCode := 0

	flag.Parse()
	runnerLogger := logger.NewRunnerLogger("")
	if *echoVersion {
		fmt.Printf("%s \n", version.GoVersion)
		fmt.Printf("Git Hash: %s \n", version.GitHash)
		fmt.Printf("Tag: %s \n", version.Tag)
		return
	}
	defer func() {
		// Just for filebeat log fetcher to differentiate following logs from others
		time.Sleep(1 * time.Second)
		log.Println("--- END OF NEBULA IMPORTER ---")
		os.Exit(errCode)
	}()

	log.Println("--- START OF NEBULA IMPORTER ---")
	if port != nil && *port > 0 && callback != nil && *callback != "" {
		// Start http server
		svr := &web.WebServer{
			Port:         *port,
			Callback:     *callback,
			RunnerLogger: runnerLogger,
		}

		if err := svr.Start(); err != nil {
			panic(err)
		}
	} else {
		if *configuration == "" {
			log.Fatal("please configure yaml file")
		}

		conf, err := config.Parse(*configuration, runnerLogger)
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
