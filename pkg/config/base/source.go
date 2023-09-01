package configbase

import (
	"io/fs"
	"os"

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

func (s *Source) Glob() ([]*Source, bool, error) {
	sourceConfig := s.SourceConfig
	src, err := sourceNew(&sourceConfig)
	if err != nil {
		return nil, false, err
	}

	g, ok := src.(source.Globber)
	if !ok {
		// Do not support glob.
		return nil, false, nil
	}
	defer src.Close()

	cs, err := g.Glob()
	if err != nil {
		return nil, true, err
	}

	if len(cs) == 0 {
		return nil, true, &os.PathError{Op: "open", Path: src.Name(), Err: fs.ErrNotExist}
	}

	ss := make([]*Source, 0, len(cs))
	for _, c := range cs {
		cpy := *s
		cpySourceConfig := c.Clone()
		cpy.SourceConfig = *cpySourceConfig
		ss = append(ss, &cpy)
	}

	return ss, true, nil
}
