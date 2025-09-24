//go:generate mockgen -source=pool.go -destination pool_mock.go -package client -aux_files github.com/vesoft-inc/nebula-importer/v4/pkg/client=client.go Pool
package client

import (
	"sync"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"github.com/cenkalti/backoff/v4"
)

type (
	Pool interface {
		Client
		GetClient(opts ...Option) (Client, error)
		ExecuteChan(statement string) (<-chan ExecuteResult, bool)
	}

	defaultPool struct {
		*options
		chExecuteDataQueue chan executeData
		lock               sync.RWMutex
		closed             bool
		done               chan struct{}
		wgSession          sync.WaitGroup
		wgStatementExecute sync.WaitGroup
	}

	NewSessionFunc func(HostAddress) Session

	executeData struct {
		statement string
		ch        chan<- ExecuteResult
	}

	ExecuteResult struct {
		Response Response
		Err      error
	}
)

func NewPool(opts ...Option) Pool {
	p := &defaultPool{
		options: newOptions(opts...),
		done:    make(chan struct{}),
	}

	p.chExecuteDataQueue = make(chan executeData, p.queueSize)

	return p
}

func (p *defaultPool) GetClient(opts ...Option) (Client, error) {
	if len(p.addresses) == 0 {
		return nil, errors.ErrNoAddresses
	}
	return p.openClient(p.addresses[0], opts...)
}

func (p *defaultPool) Open() error {
	if len(p.addresses) == 0 {
		return errors.ErrNoAddresses
	}

	for _, address := range p.addresses {
		// check if it can open successfully.
		c, err := p.openClient(address)
		if err != nil {
			return err
		}
		_ = c.Close()
	}

	p.startWorkers()

	return nil
}

func (p *defaultPool) Execute(statement string) (Response, error) {
	if p.IsClosed() {
		return nil, ErrClosed
	}
	p.wgStatementExecute.Add(1)
	defer p.wgStatementExecute.Done()

	ch := make(chan ExecuteResult, 1)
	data := executeData{
		statement: statement,
		ch:        ch,
	}
	p.chExecuteDataQueue <- data
	result := <-ch
	return result.Response, result.Err
}

func (p *defaultPool) ExecuteChan(statement string) (<-chan ExecuteResult, bool) {
	if p.IsClosed() {
		return nil, false
	}
	p.wgStatementExecute.Add(1)
	defer p.wgStatementExecute.Done()

	ch := make(chan ExecuteResult, 1)
	data := executeData{
		statement: statement,
		ch:        ch,
	}
	select {
	case p.chExecuteDataQueue <- data:
		return ch, true
	default:
		return nil, false
	}
}

func (p *defaultPool) Close() error {
	p.lock.Lock()
	p.closed = true
	p.lock.Unlock()

	p.wgStatementExecute.Wait()
	close(p.done)
	p.wgSession.Wait()
	close(p.chExecuteDataQueue)
	return nil
}

func (p *defaultPool) IsClosed() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.closed
}

func (p *defaultPool) startWorkers() {
	for _, address := range p.addresses {
		address := address
		for i := 0; i < p.concurrencyPerAddress; i++ {
			p.wgSession.Add(1)
			go func() {
				defer p.wgSession.Done()
				p.worker(address)
			}()
		}
	}
}

func (p *defaultPool) worker(address string) {
	for {
		select {
		case <-p.done:
			return
		default:
			exp := backoff.NewExponentialBackOff()
			exp.InitialInterval = p.reconnectInitialInterval
			exp.MaxInterval = DefaultReconnectMaxInterval
			exp.RandomizationFactor = DefaultRetryRandomizationFactor
			exp.Multiplier = DefaultRetryMultiplier

			var (
				err error
				c   Client
			)
			_ = backoff.Retry(func() error {
				c, err = p.openClient(address)
				if err != nil {
					p.logger.With("error", err).Error("open client failed")
				}
				return err
			}, exp)

			if err == nil {
				p.loop(c)
			}
		}
	}
}

func (p *defaultPool) openClient(address string, opts ...Option) (Client, error) {
	cloneOptions := p.options.clone()
	cloneOptions.addresses = []string{address}
	cloneOptions.withOptions(opts...)

	c := p.fnNewClientWithOptions(cloneOptions)
	if err := c.Open(); err != nil {
		return nil, err
	}

	return c, nil
}

func (p *defaultPool) loop(c Client) {
	defer func() {
		_ = c.Close()
	}()
	for {
		select {
		case data, ok := <-p.chExecuteDataQueue:
			if !ok {
				continue
			}
			resp, err := c.Execute(data.statement)
			data.ch <- ExecuteResult{
				Response: resp,
				Err:      err,
			}
		case <-p.done:
			return
		}
	}
}
