package client

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v2"
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
	pool        *nebula.ConnectionPool
	Sessions    []*nebula.Session
	requestChs  []chan base.ClientRequest
}

func NewClientPool(settings *config.NebulaClientSettings, statsCh chan<- base.Stats) (*ClientPool, error) {
	addrs := strings.Split(*settings.Connection.Address, ",")
	var hosts []nebula.HostAddress
	for _, addr := range addrs {
		hostPort := strings.Split(addr, ":")
		if len(hostPort) != 2 {
			return nil, fmt.Errorf("Invalid address: %s", addr)
		}
		port, err := strconv.Atoi(hostPort[1])
		if err != nil {
			return nil, err
		}
		hostAddr := nebula.HostAddress{Host: hostPort[0], Port: port}
		hosts = append(hosts, hostAddr)
	}
	conf := nebula.PoolConfig{
		TimeOut:         0,
		IdleTime:        0,
		MaxConnPoolSize: len(addrs) * *settings.Concurrency,
		MinConnPoolSize: 1,
	}
	connPool, err := nebula.NewConnectionPool(hosts, conf, logger.NebulaLogger{})
	if err != nil {
		return nil, err
	}
	pool := ClientPool{
		space:     *settings.Space,
		postStart: settings.PostStart,
		preStop:   settings.PreStop,
		statsCh:   statsCh,
		pool:      connPool,
	}
	pool.retry = *settings.Retry
	pool.concurrency = (*settings.Concurrency) * len(addrs)
	pool.Sessions = make([]*nebula.Session, pool.concurrency)
	pool.requestChs = make([]chan base.ClientRequest, pool.concurrency)

	j := 0
	for k := 0; k < len(addrs); k++ {
		for i := 0; i < *settings.Concurrency; i++ {
			if pool.Sessions[j], err = pool.pool.GetSession(*settings.Connection.User, *settings.Connection.Password); err != nil {
				return nil, err
			}
			pool.requestChs[j] = make(chan base.ClientRequest, *settings.ChannelBufferSize)
			j++
		}
	}

	return &pool, nil
}

func (p *ClientPool) getActiveConnIdx() int {
	for i := range p.Sessions {
		if p.Sessions[i] != nil {
			return i
		}
	}
	return -1
}

func (p *ClientPool) exec(i int, stmt string) error {
	if len(stmt) == 0 {
		return nil
	}
	resp, err := p.Sessions[i].Execute(stmt)
	if err != nil {
		return fmt.Errorf("Client(%d) fails to execute commands (%s), error: %s", i, stmt, err.Error())
	}

	if !resp.IsSucceed() {
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
		if p.Sessions[i] != nil {
			p.Sessions[i].Release()
		}
		if p.requestChs[i] != nil {
			close(p.requestChs[i])
		}
	}
	p.pool.Close()
}

func (p *ClientPool) Init() error {
	i := p.getActiveConnIdx()
	if i == -1 {
		return fmt.Errorf("no available session.")
	}
	if p.postStart != nil && p.postStart.Commands != nil {
		if err := p.exec(i, *p.postStart.Commands); err != nil {
			return err
		}
	}
	// pre-check for use space statement
	if err := p.exec(i, fmt.Sprintf("USE `%s`;", p.space)); err != nil {
		return err
	}

	for i := 0; i < p.concurrency; i++ {
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
	stmt := fmt.Sprintf("USE `%s`;", p.space)
	if err := p.exec(i, stmt); err != nil {
		logger.Error(err.Error())
		return
	}
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
		var resp *nebula.ResultSet = nil
		for retry := p.retry; retry > 0; retry-- {
			resp, err = p.Sessions[i].Execute(data.Stmt)
			if err == nil && resp.IsSucceed() {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			err = fmt.Errorf("Client %d fail to execute: %s, Error: %s", i, data.Stmt, err.Error())
		} else {
			if !resp.IsSucceed() {
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
			p.statsCh <- base.NewSuccessStats(int64(resp.GetLatency()), timeInMs, len(data.Data))
		}
	}
}
