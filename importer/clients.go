package nebula_importer

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

func InitNebulaClientPool(conf NebulaClientConfig, stmtCh <-chan Query, errLogCh chan<- error, errDataCh chan<- []interface{}) {
	for i := 0; i < conf.Concurrency; i++ {
		go func() {
			// TODO: Add retry option for graph client
			client, err := nebula.NewClient(conf.Address)
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

				for _, val := range stmt.Data {
					stmt.Stmt = strings.Replace(stmt.Stmt, "?", fmt.Sprintf("%v", val), 1)
				}

				// TODO: Add some metrics for response latency, succeededCount, failedCount
				resp, err := client.Execute(stmt.Stmt)
				if err != nil {
					errLogCh <- err
					errDataCh <- stmt.Data
					continue
				}

				if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
					errLogCh <- errors.New(fmt.Sprintf("Fail to execute: %s, error: %s", stmt, resp.GetErrorMsg()))
					errDataCh <- stmt.Data
					continue
				}
			}

		}()
	}
}
