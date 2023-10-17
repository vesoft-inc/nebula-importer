//go:generate mockgen -source=record.go -destination record_mock.go -package reader RecordReader
package reader

import (
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"
)

const (
	FormatCSV = "csv"
)

type (
	RecordReader interface {
		Source() source.Source
		source.Sizer
		Read() (int, spec.Record, error)
	}
)

func NewRecordReader(s source.Source) RecordReader {
	format := strings.ToLower(strings.TrimSpace(s.Config().Format))

	if format == "" {
		// default format
		format = FormatCSV
	}
	switch format {
	case FormatCSV:
		return NewCSVReader(s)
	// TODO: support other source formats
	default:
		panic("unsupported source format")
	}
}
