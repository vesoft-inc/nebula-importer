package client

import (
	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type NebulaClientMgr struct {
	config       *config.NebulaClientSettings
	pool         *ClientPool
	runnerLogger *logger.RunnerLogger
}

func NewNebulaClientMgr(settings *config.NebulaClientSettings, statsCh chan<- base.Stats,
	runnerLogger *logger.RunnerLogger) (*NebulaClientMgr, error) {
	mgr := NebulaClientMgr{
		config:       settings,
		runnerLogger: runnerLogger,
	}

	if pool, err := NewClientPool(settings, statsCh, runnerLogger); err != nil {
		return nil, err
	} else {
		if err := pool.Init(); err != nil {
			return nil, err
		}
		mgr.pool = pool
	}

	logger.Log.Infof("Create %d Nebula Graph clients", mgr.GetNumConnections())

	return &mgr, nil
}

func (m *NebulaClientMgr) Close() {
	m.runnerLogger.Infof("Client manager closing")
	m.pool.Close()
	m.runnerLogger.Infof("Client manager closed")
}

func (m *NebulaClientMgr) GetRequestChans() []chan base.ClientRequest {
	return m.pool.requestChs
}

func (m *NebulaClientMgr) GetNumConnections() int {
	return len(m.pool.requestChs)
}
