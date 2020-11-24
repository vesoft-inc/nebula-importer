package client

import (
	"fmt"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v2"
	"github.com/vesoft-inc/nebula-go/v2/nebula/graph"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type ClientPool struct {
	retry       int
	concurrency int
	space       string
	postStart   *config.NebulaPostStart
	preStop     *config.NebulaPreStop
	statsCh     chan<- base.Stats
	Conns       []*nebula.GraphClient
	requestChs  []chan base.ClientRequest
}

func NewClientPool(settings *config.NebulaClientSettings, statsCh chan<- base.Stats) (*ClientPool, error) {
	pool := ClientPool{
		space:     *settings.Space,
		postStart: settings.PostStart,
		preStop:   settings.PreStop,
		statsCh:   statsCh,
	}
	addrs := strings.Split(*settings.Connection.Address, ",")
	pool.retry = *settings.Retry
	pool.concurrency = (*settings.Concurrency) * len(addrs)
	pool.Conns = make([]*nebula.GraphClient, pool.concurrency)
	pool.requestChs = make([]chan base.ClientRequest, pool.concurrency)

	j := 0
	for _, addr := range addrs {
		for i := 0; i < *settings.Concurrency; i++ {
			if conn, err := NewNebulaConnection(strings.TrimSpace(addr), *settings.Connection.User, *settings.Connection.Password); err != nil {
				return nil, err
			} else {
				pool.Conns[j] = conn
				pool.requestChs[j] = make(chan base.ClientRequest, *settings.ChannelBufferSize)
				j++
			}
		}
	}

	return &pool, nil
}

func (p *ClientPool) getActiveConnIdx() int {
	for i := range p.Conns {
		if p.Conns[i] != nil {
			return i
		}
	}
	return -1
}

func (p *ClientPool) exec(i int, stmt string) error {
	if len(stmt) == 0 {
		return nil
	}
	resp, err := p.Conns[i].Execute(stmt)
	if err != nil {
		return fmt.Errorf("Client(%d) fails to execute commands (%s), error: %s", i, stmt, err.Error())
	}

	if nebula.IsError(resp) {
		return fmt.Errorf("Client(%d) fails to execute commands (%s), response error code: %v, message: %s",
			i, stmt, resp.GetErrorCode(), resp.GetErrorMsg())
	}

	return nil
}

func (p *ClientPool) Close() {
	if p.preStop != nil && p.preStop.Commands != nil {
		if i := p.getActiveConnIdx(); i != -1 {
			if err := p.exec(i, *p.preStop.Commands); err != nil {
				logger.Errorf("%s", err.Error())
			}
		}
	}

	for i := 0; i < p.concurrency; i++ {
		if p.Conns[i] != nil {
			p.Conns[i].Disconnect()
		}
		if p.requestChs[i] != nil {
			close(p.requestChs[i])
		}
	}
}

func (p *ClientPool) Init() error {
	if p.postStart != nil && p.postStart.Commands != nil {
		if i := p.getActiveConnIdx(); i != -1 {
			if err := p.exec(i, *p.postStart.Commands); err != nil {
				return err
			}
		}
	}

	stmt := fmt.Sprintf("USE `%s`;", p.space)
	for i := 0; i < p.concurrency; i++ {
		if err := p.exec(i, stmt); err != nil {
			return err
		}
		go func(i int) {
			if p.postStart != nil {
				afterPeriod, _ := time.ParseDuration(*p.postStart.AfterPeriod)
				time.Sleep(afterPeriod)
			}
			p.startWorker(i)
		}(i)
	}
	return nil
}

func (p *ClientPool) startWorker(i int) {
	for {
		data, ok := <-p.requestChs[i]
		if !ok {
			break
		}

		if data.Stmt == base.STAT_FILEDONE {
			data.ErrCh <- base.ErrData{Error: nil}
			continue
		}

		now := time.Now()

		var err error = nil
		var resp *graph.ExecutionResponse = nil
		for retry := p.retry; retry > 0; retry-- {
			resp, err = p.Conns[i].Execute(data.Stmt)
			if err == nil && !nebula.IsError(resp) {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			err = fmt.Errorf("Client %d fail to execute: %s, Error: %s", i, data.Stmt, err.Error())
		} else {
			if nebula.IsError(resp) {
				err = fmt.Errorf("Client %d fail to execute: %s, ErrMsg: %s, ErrCode: %v", i, data.Stmt, resp.GetErrorMsg(), resp.GetErrorCode())
			}
		}

		if err != nil {
			data.ErrCh <- base.ErrData{
				Error: err,
				Data:  data.Data,
			}
		} else {
			timeInMs := time.Since(now).Nanoseconds() / 1e3
			p.statsCh <- base.NewSuccessStats(int64(resp.GetLatencyInUs()), timeInMs, len(data.Data))
		}
	}
}
