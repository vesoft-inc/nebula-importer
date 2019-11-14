package client

import (
	"fmt"
	"time"

	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-go/graph"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type ClientPool struct {
	concurrency int
	space       string
	statsCh     chan<- base.Stats
	Conns       []*nebula.GraphClient
	requestChs  []chan base.ClientRequest
}

func NewClientPool(settings config.NebulaClientSettings, statsCh chan<- base.Stats) (*ClientPool, error) {
	pool := ClientPool{
		concurrency: settings.Concurrency,
		space:       settings.Space,
		statsCh:     statsCh,
	}
	pool.Conns = make([]*nebula.GraphClient, settings.Concurrency)
	pool.requestChs = make([]chan base.ClientRequest, settings.Concurrency)
	for i := 0; i < settings.Concurrency; i++ {
		if conn, err := NewNebulaConnection(settings.Connection); err != nil {
			return nil, err
		} else {
			pool.Conns[i] = conn
			chanBufferSize := 128
			if settings.ChannelBufferSize > 0 {
				chanBufferSize = settings.ChannelBufferSize
			}
			pool.requestChs[i] = make(chan base.ClientRequest, chanBufferSize)
		}
	}

	return &pool, nil
}

func (p *ClientPool) Close() {
	stmt := "UPDATE CONFIGS storage:rocksdb_column_family_options = { disable_auto_compactions = false };"
	for i := 0; i < p.concurrency; i++ {
		if p.Conns[i] != nil {
			if resp, err := p.Conns[i].Execute(stmt); err != nil {
				logger.Log.Printf("Client %d fails to open compaction option when close connection, error: %s", i, err)
			} else {
				if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
					logger.Log.Printf("Client %d fails to open compaction option when close connection, error code: %v, message: %s", i, resp.GetErrorCode(), resp.GetErrorMsg())
				}
			}
			p.Conns[i].Disconnect()
		}
		if p.requestChs[i] != nil {
			close(p.requestChs[i])
		}
	}
}

func (p *ClientPool) Init() error {
	stmt := fmt.Sprintf("USE %s; UPDATE CONFIGS storage:rocksdb_column_family_options = { disable_auto_compactions = true };", p.space)
	for i := 0; i < p.concurrency; i++ {
		if resp, err := p.Conns[i].Execute(stmt); err != nil {
			return err
		} else {
			if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
				return fmt.Errorf("Response error code: %v, message: %s", resp.GetErrorCode(), resp.GetErrorMsg())
			}
		}
		go p.startWorker(i)
	}
	return nil
}

func (p *ClientPool) startWorker(i int) {
	for {
		data, ok := <-p.requestChs[i]
		if !ok {
			break
		}

		if data.Stmt == "FILEDONE" {
			data.ErrCh <- base.ErrData{Error: nil}
			continue
		}

		now := time.Now()
		resp, err := p.Conns[i].Execute(data.Stmt)
		if err == nil && resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
			err = fmt.Errorf("Client %d fail to execute: %s, ErrMsg: %s, ErrCode: %v", i, data.Stmt, resp.GetErrorMsg(), resp.GetErrorCode())
		}

		if err != nil {
			data.ErrCh <- base.ErrData{
				Error: err,
				Data:  data.Data,
			}
		} else {
			timeInMs := time.Since(now).Microseconds()
			p.statsCh <- base.NewSuccessStats(int64(resp.GetLatencyInUs()), timeInMs, len(data.Data))
		}
	}
}
