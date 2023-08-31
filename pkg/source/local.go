package source

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	_ Source  = (*localSource)(nil)
	_ Globber = (*localSource)(nil)
)

type (
	LocalConfig struct {
		Path string `yaml:"path,omitempty"`
	}

	localSource struct {
		c *Config
		f *os.File
	}
)

func newLocalSource(c *Config) Source {
	return &localSource{
		c: c,
	}
}

func (s *localSource) Name() string {
	return s.c.Local.String()
}

func (s *localSource) Glob() ([]*Config, error) {
	matches, err := filepath.Glob(s.c.Local.Path)
	if err != nil {
		return nil, err
	}

	cs := make([]*Config, 0, len(matches))
	for _, match := range matches {
		cpy := s.c.Clone()
		cpy.Local.Path = match
		cs = append(cs, cpy)
	}
	return cs, nil
}

func (s *localSource) Open() error {
	f, err := os.Open(s.c.Local.Path)
	if err != nil {
		return err
	}
	s.f = f
	return nil
}

func (s *localSource) Config() *Config {
	return s.c
}

func (s *localSource) Size() (int64, error) {
	fi, err := s.f.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (s *localSource) Read(p []byte) (int, error) {
	return s.f.Read(p)
}

func (s *localSource) Close() (err error) {
	if s.f != nil {
		err = s.f.Close()
	}
	return err
}

func (c *LocalConfig) String() string {
	return fmt.Sprintf("local %s", c.Path)
}
