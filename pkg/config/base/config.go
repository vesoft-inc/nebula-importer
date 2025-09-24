package configbase

import (
	"log/slog"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
)

type Configurator interface {
	Optimize(configPath string) error
	Build(l *slog.Logger) error
	GetLogger() *slog.Logger
	GetClientPool() client.Pool
	GetManager() manager.Manager
}
