//go:build !cgo || !libhdfs

package source

import (
	"errors"
	"fmt"
)

var errLibHDFSUnavailable = errors.New("the libhdfs backend requires a build with CGO_ENABLED=1 and -tags libhdfs")

type libhdfsSource struct {
	c *Config
}

func newLibHDFSSource(c *Config) Source {
	return &libhdfsSource{c: c}
}

func (s *libhdfsSource) Name() string {
	return s.c.HDFS.String()
}

func (s *libhdfsSource) Open() error {
	return s.unavailable()
}

func (s *libhdfsSource) Glob() ([]*Config, error) {
	return nil, s.unavailable()
}

func (s *libhdfsSource) Config() *Config {
	return s.c
}

func (s *libhdfsSource) Size() (int64, error) {
	return 0, s.unavailable()
}

func (s *libhdfsSource) Read(_ []byte) (int, error) {
	return 0, s.unavailable()
}

func (*libhdfsSource) Close() error {
	return nil
}

func (s *libhdfsSource) IsDir(_ string) (bool, error) {
	return false, s.unavailable()
}

func (s *libhdfsSource) Readdirnames(_ string) ([]string, error) {
	return nil, s.unavailable()
}

func (s *libhdfsSource) unavailable() error {
	return fmt.Errorf("%w: %s", errLibHDFSUnavailable, s.Name())
}
