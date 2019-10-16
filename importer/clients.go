package nebula_importer

import (
	"errors"
	"fmt"
	"log"

	nebula "github.com/vesoft-inc/nebula-go"
	graph "github.com/vesoft-inc/nebula-go/graph"
)

type NebulaClientConfig struct {
	Address     string
	Retry       int
	Concurrency int
	User        string
	Password    string
}

func InitNebulaClientPool(conf NebulaClientConfig, stmtCh <-chan string, errCh chan<- error) {
	for i := 0; i < conf.Concurrency; i++ {
		go func() {
			// TODO: Add retry option for graph client
			client, err := nebula.NewClient(conf.Address, nebula.GraphOptions{})
			if err != nil {
				log.Println(err)
				return
			}

			if err = client.Connect(conf.User, conf.Password); err != nil {
				log.Println(err)
				return
			}
			defer client.Disconnect()

			for {
				stmt := <-stmtCh

				// TODO: Add some metrics for response latency, succeededCount, failedCount
				resp, err := client.Execute(stmt)
				if err != nil {
					errCh <- err
					continue
				}

				if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
					errCh <- errors.New(fmt.Sprintf("Fail to execute: %s, error: %s", stmt, resp.GetErrorMsg()))
					continue
				}
			}

		}()
	}
}
