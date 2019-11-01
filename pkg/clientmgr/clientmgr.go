package clientmgr

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go"

	"github.com/vesoft-inc/nebula-go/graph"
	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
)

type NebulaClientMgr struct {
	Config  config.NebulaClientSettings
	ErrCh   chan<- base.ErrData
	StatsCh chan<- stats.Stats
	DoneCh  <-chan bool
}

func (m *NebulaClientMgr) InitNebulaClientPool() []chan base.Stmt {
	stmtChs := make([]chan base.Stmt, m.Config.Concurrency)
	for i := 0; i < m.Config.Concurrency; i++ {
		stmtChs[i] = make(chan base.Stmt)
	}

	for i := 0; i < m.Config.Concurrency; i++ {
		go func(i int) {
			// TODO: Add retry option for graph client
			client, err := nebula.NewClient(m.Config.Connection.Address)
			if err != nil {
				log.Println(err)
				return
			}

			if err = client.Connect(m.Config.Connection.User, m.Config.Connection.Password); err != nil {
				log.Println(err)
				return
			}
			defer client.Disconnect()

			for {
				select {
				case <-m.DoneCh:
					m.ErrCh <- base.ErrData{Done: true}
				case stmt := <-stmtChs[i]:
					for _, val := range stmt.Data {
						stmt.Stmt = strings.Replace(stmt.Stmt, "?", fmt.Sprintf("%v", val), 1)
					}

					// TODO: Add some metrics for response latency, succeededCount, failedCount
					now := time.Now()
					resp, err := client.Execute(stmt.Stmt)
					reqTime := time.Since(now).Seconds()

					if err != nil {
						m.ErrCh <- base.ErrData{
							Error: err,
							Data:  stmt.Data,
							Done:  false,
						}
						continue
					}

					if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
						errMsg := fmt.Sprintf("Fail to execute: %s, ErrMsg: %s, ErrCode: %v", stmt.Stmt, resp.GetErrorMsg(), resp.GetErrorCode())
						m.ErrCh <- base.ErrData{
							Error: errors.New(errMsg),
							Data:  stmt.Data,
							Done:  false,
						}
						continue
					}
					m.StatsCh <- stats.Stats{
						Latency: uint64(resp.GetLatencyInUs()),
						ReqTime: reqTime,
					}
				}
			}
		}(i)
	}
	log.Printf("Create %d Nebula Graph clients", m.Config.Concurrency)
	return stmtChs
}
