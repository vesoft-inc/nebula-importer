package web

import (
	"sync"

	"github.com/vesoft-inc/nebula-importer/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type taskMgr struct {
	tasks map[uint64]*cmd.Runner
	mux   sync.Mutex
}

func (m *taskMgr) put(k uint64, r *cmd.Runner) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.tasks[k] = r
}

func (m *taskMgr) get(k uint64) *cmd.Runner {
	m.mux.Lock()
	defer m.mux.Unlock()
	if v, ok := m.tasks[k]; !ok {
		logger.Errorf("Fail to get %s value from task manager", k)
		return nil
	} else {
		return v
	}
}

func (m *taskMgr) del(k uint64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.tasks, k)
}
