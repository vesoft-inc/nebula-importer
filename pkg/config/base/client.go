package configbase

import (
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
)

var newClientPool = client.NewPool

const (
	ClientVersion3       = "v3"
	ClientVersionDefault = ClientVersion3
)

type Client struct {
	Version                  string        `yaml:"version"`
	Address                  string        `yaml:"address"`
	User                     string        `yaml:"user,omitempty"`
	Password                 string        `yaml:"password,omitempty"`
	ConcurrencyPerAddress    int           `yaml:"concurrencyPerAddress,omitempty"`
	ReconnectInitialInterval time.Duration `yaml:"reconnectInitialInterval,omitempty"`
	Retry                    int           `yaml:"retry,omitempty"`
	RetryInitialInterval     time.Duration `yaml:"retryInitialInterval,omitempty"`
}

func (c *Client) BuildClientPool(opts ...client.Option) (client.Pool, error) {
	if c.Version == "" {
		c.Version = ClientVersion3
	}
	options := make([]client.Option, 0, 7+len(opts))
	options = append(
		options,
		client.WithAddress(c.Address),
		client.WithUserPassword(c.User, c.Password),
		client.WithReconnectInitialInterval(c.ReconnectInitialInterval),
		client.WithRetry(c.Retry),
		client.WithRetryInitialInterval(c.RetryInitialInterval),
		client.WithConcurrencyPerAddress(c.ConcurrencyPerAddress),
	)
	switch c.Version {
	case ClientVersion3:
		options = append(options, client.WithV3())
	default:
		return nil, errors.ErrUnsupportedClientVersion
	}
	options = append(options, opts...)
	pool := newClientPool(options...)
	return pool, nil
}
