package client

import (
	"fmt"
	"time"

	"github.com/vesoft-inc/nebula-go/graph"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type NebulaClientMgr struct {
	config config.NebulaClientSettings
	pool   *ClientPool
}

func NewNebulaClientMgr(settings config.NebulaClientSettings) (*NebulaClientMgr, error) {
	mgr := NebulaClientMgr{
		config: settings,
	}

	if pool, err := NewClientPool(settings); err != nil {
		return nil, err
	} else {
		mgr.pool = pool
	}

	for i := 0; i < settings.Concurrency; i++ {
		go mgr.startWorker(i)
	}

	logger.Log.Printf("Create %d Nebula Graph clients", mgr.config.Concurrency)

	return &mgr, nil
}

func (m *NebulaClientMgr) Close() {
	m.pool.Close()
}

func (m *NebulaClientMgr) GetDataChans() []chan base.ClientRequest {
	return m.pool.DataChs
}

func (m *NebulaClientMgr) startWorker(i int) {
	for {
		data, ok := <-m.pool.DataChs[i]
		if !ok {
			break
		}

		now := time.Now()
		resp, err := m.pool.Conns[i].Execute(data.Stmt)
		if err == nil && resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
			err = fmt.Errorf("Client %d fail to execute: %s, ErrMsg: %s, ErrCode: %v", i, data.Stmt, resp.GetErrorMsg(), resp.GetErrorCode())
		}

		if err != nil {
			data.ResponseCh <- base.ResponseData{Error: err}
		} else {
			data.ResponseCh <- base.ResponseData{
				Error: nil,
				Stats: base.Stats{
					Latency: uint64(resp.GetLatencyInUs()),
					ReqTime: time.Since(now).Seconds(),
				},
			}
		}
	}
}
