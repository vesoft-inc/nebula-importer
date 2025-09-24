package configv3

import (
	"log/slog"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	configbase "github.com/vesoft-inc/nebula-importer/v4/pkg/config/base"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/reader"
)

type (
	Manager struct {
		GraphName          string `yaml:"spaceName"`
		configbase.Manager `yaml:",inline"`
	}
)

func (m *Manager) BuildManager(
	l *slog.Logger,
	pool client.Pool,
	sources Sources,
	opts ...manager.Option,
) (manager.Manager, error) {
	options := make([]manager.Option, 0, 8+len(opts))
	options = append(options,
		manager.WithClientPool(pool),
		manager.WithBatch(m.Batch),
		manager.WithReaderConcurrency(m.ReaderConcurrency),
		manager.WithImporterConcurrency(m.ImporterConcurrency),
		manager.WithStatsInterval(m.StatsInterval),
		manager.WithBeforeHooks(m.Hooks.Before...),
		manager.WithAfterHooks(m.Hooks.After...),
		manager.WithLogger(l),
	)
	options = append(options, opts...)

	mgr := manager.NewWithOpts(options...)

	for i := range sources {
		s := sources[i]
		src, brr, err := s.BuildSourceAndReader(reader.WithBatch(m.Batch), reader.WithLogger(l))
		if err != nil {
			return nil, err
		}

		importers, err := s.BuildImporters(m.GraphName, pool)
		if err != nil {
			return nil, err
		}
		if err = mgr.Import(src, brr, importers...); err != nil {
			return nil, err
		}
	}

	return mgr, nil
}
