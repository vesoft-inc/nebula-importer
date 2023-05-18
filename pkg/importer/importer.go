//go:generate mockgen -source=importer.go -destination importer_mock.go -package importer Importer
package importer

import (
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"
)

type (
	Importer interface {
		Import(records ...spec.Record) (*ImportResp, error)

		// Add Done Wait for synchronize, similar to sync.WaitGroup.
		Add(delta int)
		Done()
		Wait()
	}

	ImportResp struct {
		Latency  time.Duration
		RespTime time.Duration
	}

	ImportResult struct {
		Resp *ImportResp
		Err  error
	}

	Option func(*defaultImporter)

	defaultImporter struct {
		builder spec.StatementBuilder
		pool    client.Pool

		fnAdd  func(delta int)
		fnDone func()
		fnWait func()
	}
)

func New(builder spec.StatementBuilder, pool client.Pool, opts ...Option) Importer {
	options := make([]Option, 0, 2+len(opts))
	options = append(options, WithStatementBuilder(builder), WithClientPool(pool))
	options = append(options, opts...)
	return NewWithOpts(options...)
}

func NewWithOpts(opts ...Option) Importer {
	i := &defaultImporter{}
	for _, opt := range opts {
		opt(i)
	}
	if i.fnAdd == nil {
		i.fnAdd = func(delta int) {}
	}
	if i.fnDone == nil {
		i.fnDone = func() {}
	}
	if i.fnWait == nil {
		i.fnWait = func() {}
	}
	return i
}

func WithStatementBuilder(builder spec.StatementBuilder) Option {
	return func(i *defaultImporter) {
		i.builder = builder
	}
}

func WithClientPool(p client.Pool) Option {
	return func(i *defaultImporter) {
		i.pool = p
	}
}

func WithAddFunc(fn func(delta int)) Option {
	return func(i *defaultImporter) {
		i.fnAdd = fn
	}
}

func WithDoneFunc(fn func()) Option {
	return func(i *defaultImporter) {
		i.fnDone = fn
	}
}

func WithWaitFunc(fn func()) Option {
	return func(i *defaultImporter) {
		i.fnWait = fn
	}
}

func (i *defaultImporter) Import(records ...spec.Record) (*ImportResp, error) {
	statement, err := i.builder.Build(records...)
	if err != nil {
		return nil, err
	}

	resp, err := i.pool.Execute(statement)
	if err != nil {
		return nil, errors.NewImportError(err).
			SetStatement(statement)
	}
	if !resp.IsSucceed() {
		return nil, errors.NewImportError(err, "the execute error is %s ", resp.GetError()).
			SetStatement(statement)
	}

	return &ImportResp{
		RespTime: resp.GetRespTime(),
		Latency:  resp.GetLatency(),
	}, nil
}

func (i *defaultImporter) Add(delta int) {
	i.fnAdd(delta)
}

func (i *defaultImporter) Done() {
	i.fnDone()
}

func (i *defaultImporter) Wait() {
	i.fnWait()
}
