package nebula_importer

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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

type NebulaClientMgr struct {
	ErrCh   chan<- ErrData
	StatsCh chan<- Stats
	DoneCh  <-chan bool
}

func (m *NebulaClientMgr) InitNebulaClientPool(conf NebulaClientConfig) []chan Stmt {
	stmtChs := make([]chan Stmt, conf.Concurrency)
	for i := 0; i < conf.Concurrency; i++ {
		stmtChs[i] = make(chan Stmt)
	}

	for i := 0; i < conf.Concurrency; i++ {
		go func(i int) {
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
				select {
				case <-m.DoneCh:
					m.ErrCh <- ErrData{Done: true}
				case stmt := <-stmtChs[i]:
					for _, val := range stmt.Data {
						stmt.Stmt = strings.Replace(stmt.Stmt, "?", fmt.Sprintf("%v", val), 1)
					}

					// TODO: Add some metrics for response latency, succeededCount, failedCount
					now := time.Now()
					resp, err := client.Execute(stmt.Stmt)
					reqTime := time.Since(now).Seconds()

					if err != nil {
						m.ErrCh <- ErrData{
							Error: err,
							Data:  stmt.Data,
							Done:  false,
						}
						continue
					}

					if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
						m.ErrCh <- ErrData{
							Error: errors.New(fmt.Sprintf("Fail to execute: %s, ErrMsg: %s, ErrCode: %v", stmt.Stmt, resp.GetErrorMsg(), resp.GetErrorCode())),
							Data:  stmt.Data,
							Done:  false,
						}
						continue
					}

					m.StatsCh <- Stats{
						Latency: uint64(resp.GetLatencyInUs()),
						ReqTime: reqTime,
						Done:    false,
					}
				}
			}

		}(i)
	}
	log.Printf("Create %d Nebula Graph clients", conf.Concurrency)
	return stmtChs
}
