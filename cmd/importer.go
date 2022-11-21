package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/version"
	"github.com/vesoft-inc/nebula-importer/pkg/web"
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

	paramMap := make(map[string]string)
	if flag.NArg() > 0 {
		for i := 0; i != flag.NArg(); i++ {
			s := flag.Arg(i)
			sa := strings.Split(s, "=")
			if len(sa) == 2 {
				paramMap[sa[0]] = sa[1]
				log.Printf("=== External Param args[%d] is [%s]", i, s)
			}
		}
	}

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

		fileArr := conf.Files
		applyExParam := make(map[string]string)
		for _, value := range fileArr {
			if strings.ToUpper(*value.Schema.Type) == "VERTEX" {
				tagArr := value.Schema.Vertex.Tags
				if len(tagArr) > 0 {
					for _, tag := range tagArr {
						tagName := tag.Name
						exVertexPath, isOk := paramMap[*tagName]
						if isOk {
							*value.Path = exVertexPath
							log.Printf("=== Update tag [%s] conf, new path is %s.", *tagName, *value.Path)

							applyExParam[*tagName] = *value.Path
						}
					}
				}
			} else {
				edgeName := value.Schema.Edge.Name
				exEdgePath, isOk := paramMap[*edgeName]
				if isOk {
					*value.Path = exEdgePath
					log.Printf("=== Update edge [%s] conf, new path is %s.", *edgeName, *value.Path)

					applyExParam[*edgeName] = *value.Path
				}
			}
		}

		if len(applyExParam) > 0 {
			log.Println("--- START OF Update Conf YAML ---")
			for k, v := range applyExParam {
				log.Printf("=== External Apply Param key[%s] is [%s]", k, v)
			}
			err := config.UpdateParse(*configuration, conf)
			if err != nil {
				e := err.(errors.ImporterError)
				log.Println(e.ErrMsg.Error())
				errCode = e.ErrCode
				return
			}
			log.Println("--- END OF Update Conf YAML ---")
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
