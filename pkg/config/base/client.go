package configbase

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/manager"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/utils"
)

var newClientPool = client.NewPool

const (
	ClientVersion3       = "v3"
	ClientVersionDefault = ClientVersion3
)

type (
	Client struct {
		Version                  string        `yaml:"version"`
		Address                  string        `yaml:"address"`
		User                     string        `yaml:"user,omitempty"`
		Password                 string        `yaml:"password,omitempty"`
		ConcurrencyPerAddress    int           `yaml:"concurrencyPerAddress,omitempty"`
		ReconnectInitialInterval time.Duration `yaml:"reconnectInitialInterval,omitempty"`
		Retry                    int           `yaml:"retry,omitempty"`
		RetryInitialInterval     time.Duration `yaml:"retryInitialInterval,omitempty"`
		SSL                      *SSL          `yaml:"ssl,omitempty"`
	}

	SSL struct {
		Enable             bool   `yaml:"enable,omitempty"`
		CertPath           string `yaml:"certPath,omitempty"`
		KeyPath            string `yaml:"keyPath,omitempty"`
		CAPath             string `yaml:"caPath,omitempty"`
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty"`
	}
)

// OptimizePath optimizes relative paths base to the configuration file path
func (c *Client) OptimizePath(configPath string) error {
	if c == nil {
		return nil
	}

	if c.SSL != nil && c.SSL.Enable {
		configPathDir := filepath.Dir(configPath)
		c.SSL.CertPath = utils.RelativePathBaseOn(configPathDir, c.SSL.CertPath)
		c.SSL.KeyPath = utils.RelativePathBaseOn(configPathDir, c.SSL.KeyPath)
		c.SSL.CAPath = utils.RelativePathBaseOn(configPathDir, c.SSL.CAPath)
	}

	return nil
}

func (c *Client) BuildClientPool(opts ...client.Option) (client.Pool, error) {
	if c.Version == "" {
		c.Version = ClientVersion3
	}
	tlsConfig, err := c.SSL.BuildConfig()
	if err != nil {
		return nil, err
	}

	options := make([]client.Option, 0, 8+len(opts))
	options = append(
		options,
		client.WithAddress(c.Address),
		client.WithUserPassword(c.User, c.Password),
		client.WithTLSConfig(tlsConfig),
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
	sessionPool, err := client.NewSessionPool(options...)
	if err != nil {
		return nil, err
	}
	manager.DefaultSessionPool = sessionPool
	return pool, nil
}

func (s *SSL) BuildConfig() (*tls.Config, error) {
	if s == nil || !s.Enable {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		Certificates:       make([]tls.Certificate, 1),
		InsecureSkipVerify: s.InsecureSkipVerify, //nolint:gosec
	}

	rootPEM, err := os.ReadFile(s.CAPath)
	if err != nil {
		return nil, err
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(rootPEM)
	tlsConfig.RootCAs = rootCAs

	cert, err := tls.LoadX509KeyPair(s.CertPath, s.KeyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates[0] = cert

	return tlsConfig, nil
}
