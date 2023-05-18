package configbase

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
)

type Configurator interface {
	Optimize(configPath string) error
	Build() error
	GetLogger() logger.Logger
	GetClientPool() client.Pool
	GetManager() manager.Manager
}
