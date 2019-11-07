package client

import (
	"log"

	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
)

type ClientPool struct {
	concurrency int
	Conns       []*nebula.GraphClient
	DataChs     []chan base.Data
}

func NewClientPool(settings config.NebulaClientSettings) *ClientPool {
	pool := ClientPool{
		concurrency: settings.Concurrency,
	}
	pool.Conns = make([]*nebula.GraphClient, settings.Concurrency)
	pool.DataChs = make([]chan base.Data, settings.Concurrency)
	for i := 0; i < settings.Concurrency; i++ {
		if conn, err := NewNebulaConnection(settings.Connection); err != nil {
			log.Fatal("Fail to create client pool, ", err.Error())
		} else {
			pool.Conns[i] = conn
			pool.DataChs[i] = make(chan base.Data, 128)
		}
	}

	return &pool
}

func (p *ClientPool) Close() {
	for i := 0; i < p.concurrency; i++ {
		p.Conns[i].Disconnect()
		close(p.DataChs[i])
	}
}
