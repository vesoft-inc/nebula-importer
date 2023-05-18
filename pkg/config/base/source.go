package configbase

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/reader"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
)

var sourceNew = source.New

type (
	Source struct {
		SourceConfig source.Config `yaml:",inline"`
		Batch        int           `yaml:"batch,omitempty"`
	}
)

func (s *Source) BuildSourceAndReader(opts ...reader.Option) (
	source.Source,
	reader.BatchRecordReader,
	error,
) {
	sourceConfig := s.SourceConfig
	src, err := sourceNew(&sourceConfig)
	if err != nil {
		return nil, nil, err
	}
	if s.Batch > 0 {
		// Override the batch in the manager.
		opts = append(opts, reader.WithBatch(s.Batch))
	}

	rr := reader.NewRecordReader(src)
	brr := reader.NewBatchRecordReader(rr, opts...)
	return src, brr, nil
}
