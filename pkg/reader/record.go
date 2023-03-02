//go:generate mockgen -source=record.go -destination record_mock.go -package reader RecordReader
package reader

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"
)

type (
	RecordReader interface {
		Source() source.Source
		source.Sizer
		Read() (int, spec.Record, error)
	}
)

func NewRecordReader(s source.Source) RecordReader {
	// TODO: support other source formats
	return NewCSVReader(s)
}
