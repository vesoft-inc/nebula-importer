//go:generate mockgen -source=client.go -destination client_mock.go -package client Client
package client

import (
	"strconv"
	"strings"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"github.com/cenkalti/backoff/v4"
	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type (
	Client interface {
		Open() error
		Execute(statement string) (Response, error)
		Close() error
	}

	HostAddress struct {
		Host string
		Port int
	}

	defaultClient struct {
		*options
		session Session
	}
)

func NewClient(opts ...Option) Client {
	return newClientWithOptions(newOptions(opts...))
}

func newClientWithOptions(o *options) Client {
	return &defaultClient{
		options: o,
	}
}

func (c *defaultClient) Open() error {
	if len(c.addresses) == 0 {
		return errors.ErrNoAddresses
	}
	hostPort := strings.Split(c.addresses[0], ":")
	if len(hostPort) != 2 {
		return errors.ErrInvalidAddress
	}
	if hostPort[0] == "" {
		return errors.ErrInvalidAddress
	}
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		return errors.ErrInvalidAddress
	}
	hostAddress := HostAddress{Host: hostPort[0], Port: port}

	session := c.fnNewSession(hostAddress)
	if err = session.Open(); err != nil {
		return err
	}

	c.session = session

	if c.clientInitFunc != nil {
		if err = c.clientInitFunc(c); err != nil {
			c.session = nil
			_ = session.Close()
			return err
		}
	}

	return nil
}

func (c *defaultClient) Execute(statement string) (Response, error) {
	exp := backoff.NewExponentialBackOff()
	exp.InitialInterval = c.retryInitialInterval
	exp.MaxInterval = DefaultRetryMaxInterval
	exp.MaxElapsedTime = DefaultRetryMaxElapsedTime
	exp.Multiplier = DefaultRetryMultiplier
	exp.RandomizationFactor = DefaultRetryRandomizationFactor

	var (
		err   error
		resp  Response
		retry = c.retry
	)

	// There are three cases of retry
	// * Case 1: retry no more
	// * Case 2. retry as much as possible
	// * Case 3: retry with limit times
	_ = backoff.Retry(func() error {
		resp, err = c.session.Execute(statement)
		if err == nil && resp.IsSucceed() {
			return nil
		}
		retryErr := err
		if resp != nil {
			retryErr = resp.GetError()

			// Case 1: retry no more
			if resp.IsPermanentError() {
				// stop the retry
				return backoff.Permanent(retryErr)
			}

			// Case 2. retry as much as possible
			if resp.IsRetryMoreError() {
				retry = c.retry
				return retryErr
			}
		}

		// Case 3: retry with limit times
		if retry <= 0 {
			// stop the retry
			return backoff.Permanent(retryErr)
		}
		retry--
		return retryErr
	}, exp)
	if err != nil {
		c.logger.WithError(err).Error("execute statement failed")
	}
	return resp, err
}

func (c *defaultClient) Close() error {
	if c.session != nil {
		if err := c.session.Close(); err != nil {
			return err
		}
		c.session = nil
	}
	return nil
}

func NewSessionPool(opts ...Option) (*nebula.SessionPool, error) {
	ops := newOptions(opts...)
	var (
		hostAddresses []nebula.HostAddress
		pool          *nebula.SessionPool
	)

	for _, h := range ops.addresses {
		hostPort := strings.Split(h, ":")
		if len(hostPort) != 2 {
			return nil, errors.ErrInvalidAddress
		}
		if hostPort[0] == "" {
			return nil, errors.ErrInvalidAddress
		}
		port, err := strconv.Atoi(hostPort[1])
		if err != nil {
			err = errors.ErrInvalidAddress
		}
		hostAddresses = append(hostAddresses, nebula.HostAddress{Host: hostPort[0], Port: port})
	}
	conf, err := nebula.NewSessionPoolConf(ops.user, ops.password, hostAddresses,
		"sf300_2", nebula.WithMaxSize(3000))
	if err != nil {
		return nil, err
	}
	pool, err = nebula.NewSessionPool(*conf, nebula.DefaultLogger{})
	if err != nil {
		return nil, err
	}
	return pool, nil

}
