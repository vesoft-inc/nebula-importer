package client

import (
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type NebulaClientMgr struct {
	config *config.NebulaClientSettings
	pool   *ClientPool
}

func NewNebulaClientMgr(settings *config.NebulaClientSettings, statsCh chan<- base.Stats) (*NebulaClientMgr, error) {
	mgr := NebulaClientMgr{
		config: settings,
	}

	if pool, err := NewClientPool(settings, statsCh); err != nil {
		return nil, err
	} else {
		if err := pool.Init(); err != nil {
			return nil, err
		}
		mgr.pool = pool
	}

	logger.Infof("Create %d Nebula Graph clients", mgr.GetNumConnections())

	return &mgr, nil
}

func (m *NebulaClientMgr) Close() {
	m.pool.Close()
}

func (m *NebulaClientMgr) GetRequestChans() []chan base.ClientRequest {
	return m.pool.requestChs
}

func (m *NebulaClientMgr) GetNumConnections() int {
	return len(m.pool.requestChs)
}
