package importer

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go"
	graph "github.com/vesoft-inc/nebula-go/graph"
)

type NebulaClientMgr struct {
	Config  Settings
	ErrCh   chan<- ErrData
	StatsCh chan<- Stats
	DoneCh  <-chan bool
}

func (m *NebulaClientMgr) InitNebulaClientPool() []chan Stmt {
	stmtChs := make([]chan Stmt, m.Config.Concurrency)
	for i := 0; i < m.Config.Concurrency; i++ {
		stmtChs[i] = make(chan Stmt)
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
					}
				}
			}
		}(i)
	}
	log.Printf("Create %d Nebula Graph clients", m.Config.Concurrency)
	return stmtChs
}
