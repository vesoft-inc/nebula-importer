package clientmgr

import (
	"log"

	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/yixinglu/nebula-importer/pkg/config"
)

type Record []string

type ClientPool struct {
	concurrency int
	Conns       []*nebula.GraphClient
	DataChs     []chan Record
}

func DoneRecord() {
	return Record{"DONE"}
}

func NewClientPool(settings config.NebulaClientSettings) *ClientPool {
	pool := ClientPool{
		concurrency: settings.Concurrency,
	}
	pool.Conns = make([]*nebula.GraphClient, settings.Concurrency)
	pool.DataChs = make([]chan Record, settings.Concurrency)
	for i := 0; i < settings.Concurrency; i++ {
		pool.Conns[i], err = NewNebulaClient(settings.Connection)
		if err != nil {
			log.Println("Fail to create client pool, ", err.Error())
			continue
		}
		pool.DataChs[i] = make(chan Record)
	}

	return &pool
}

func (p *ClientPool) Close() {
	for i := 0; i < p.concurrency; i++ {
		p.Conns[i].Disconnect()
		close(p.DataChs[i])
	}
}
