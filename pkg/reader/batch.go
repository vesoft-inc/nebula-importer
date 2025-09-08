//go:generate mockgen -source=batch.go -destination batch_mock.go -package reader BatchRecordReader
package reader

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"
)

type (
	BatchRecordReader interface {
		Source() source.Source
		source.Sizer
		ReadBatch() (int, spec.Records, error)
	}

	continueError struct {
		Err error
	}

	defaultBatchReader struct {
		*options
		rr RecordReader
	}
)

func NewBatchRecordReader(rr RecordReader, opts ...Option) BatchRecordReader {
	brr := &defaultBatchReader{
		options: newOptions(opts...),
		rr:      rr,
	}
	brr.logger = brr.logger.With(logger.Field{Key: "source", Value: rr.Source().Name()})
	return brr
}

func NewContinueError(err error) error {
	return &continueError{
		Err: err,
	}
}

func (r *defaultBatchReader) Source() source.Source {
	return r.rr.Source()
}

func (r *defaultBatchReader) Size() (int64, error) {
	return r.rr.Size()
}

func (r *defaultBatchReader) ReadBatch() (int, spec.Records, error) { //nolint:gocritic
	var (
		totalBytes int
		records    = make(spec.Records, 0, r.batch)
	)

	for batch := 0; batch < r.batch; {
		n, record, err := r.rr.Read()
		totalBytes += n
		if err != nil {
			// case1: Read continue error.
			if ce := new(continueError); stderrors.As(err, &ce) {
				r.logger.WithError(ce.Err).Error("read source failed")
				continue
			}

			// case2: Read error and still have records.
			if totalBytes > 0 {
				break
			}

			// Read error and have no records.
			return 0, nil, err
		}
		batch++
		records = append(records, record)
	}
	return totalBytes, records, nil
}

func (ce *continueError) Error() string {
	return ce.Err.Error()
}

func (ce *continueError) Cause() error {
	return ce.Err
}

func (ce *continueError) Unwrap() error {
	return ce.Err
}
