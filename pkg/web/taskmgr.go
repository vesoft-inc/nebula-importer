package web

import (
	"sync"

	"github.com/vesoft-inc/nebula-importer/v3/pkg/cmd"
	"github.com/vesoft-inc/nebula-importer/v3/pkg/logger"
)

type taskMgr struct {
	tasks map[string]*cmd.Runner
	mux   sync.Mutex
}

func newTaskMgr() *taskMgr {
	return &taskMgr{
		tasks: make(map[string]*cmd.Runner),
	}
}

func (m *taskMgr) keys() []string {
	m.mux.Lock()
	defer m.mux.Unlock()
	var keys []string
	for k := range m.tasks {
		keys = append(keys, k)
	}
	return keys
}

func (m *taskMgr) put(k string, r *cmd.Runner) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.tasks[k] = r
}

func (m *taskMgr) get(k string) *cmd.Runner {
	m.mux.Lock()
	defer m.mux.Unlock()
	if v, ok := m.tasks[k]; !ok {
		logger.Log.Errorf("Fail to get %s value from task manager", k)
		return nil
	} else {
		return v
	}
}

func (m *taskMgr) del(k string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.tasks, k)
}
