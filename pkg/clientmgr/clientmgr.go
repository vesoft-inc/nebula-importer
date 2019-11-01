package clientmgr

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vesoft-inc/nebula-go/graph"
	"github.com/yixinglu/nebula-importer/pkg/base"
	"github.com/yixinglu/nebula-importer/pkg/config"
	"github.com/yixinglu/nebula-importer/pkg/stats"
)

type NebulaClientMgr struct {
	config  config.NebulaClientSettings
	file    config.File
	errCh   chan<- base.ErrData
	statsCh chan<- stats.Stats
	doneCh  <-chan bool
}

func NewNebulaClientMgr(settings config.NebulaClientSettings, file config.File) *NebulaClientMgr {
	return &NebulaClientMgr{
		config: settings,
		file:   file,
	}
}

func (m *NebulaClientMgr) withErrChan(errCh chan<- base.ErrData) *NebulaClientMgr {
	m.errCh = errCh
	return m
}

func (m *NebulaClientMgr) withStatsChan(statsCh chan<- stats.Stats) *NebulaClientMgr {
	m.statsCh = statsCh
	return m
}

func (m *NebulaClientMgr) withDoneCh(doneCh chan<- bool) *NebulaClientMgr {
	m.doneCh = doneCh
	return m
}

func (m *NebulaClientMgr) InitNebulaClientPool() []chan base.Stmt {
	stmtChs := make([]chan base.Stmt, m.Config.Concurrency)
	for i := 0; i < m.Config.Concurrency; i++ {
		stmtChs[i] = make(chan base.Stmt)
	}

	for i := 0; i < m.Config.Concurrency; i++ {
		go func(i int) {
			// TODO: Add retry option for graph client
			client, err := NewClientPool(m.config.Connection)
			if err != nil {
				log.Println("Fail to create client pool, ", err.Error())
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
