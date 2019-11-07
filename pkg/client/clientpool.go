package client

import (
	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
)

type ClientPool struct {
	concurrency int
	Conns       []*nebula.GraphClient
	DataChs     []chan base.Data
}

func NewClientPool(settings config.NebulaClientSettings) (*ClientPool, error) {
	pool := ClientPool{
		concurrency: settings.Concurrency,
	}
	pool.Conns = make([]*nebula.GraphClient, settings.Concurrency)
	pool.DataChs = make([]chan base.Data, settings.Concurrency)
	for i := 0; i < settings.Concurrency; i++ {
		if conn, err := NewNebulaConnection(settings.Connection); err != nil {
			return nil, err
		} else {
			pool.Conns[i] = conn
			chanBufferSize := 128
			if settings.ChannelBufferSize > 0 {
				chanBufferSize = settings.ChannelBufferSize
			}
			pool.DataChs[i] = make(chan base.Data, chanBufferSize)
		}
	}

	return &pool, nil
}

func (p *ClientPool) Close() {
	for i := 0; i < p.concurrency; i++ {
		if p.Conns[i] != nil {
			p.Conns[i].Disconnect()
		}
		if p.DataChs[i] != nil {
			close(p.DataChs[i])
		}
	}
}
