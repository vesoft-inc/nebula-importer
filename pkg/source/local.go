package source

import (
	"fmt"
	"os"
)

var _ Source = (*localSource)(nil)

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

func (s *localSource) Close() error {
	return s.f.Close()
}

func (c *LocalConfig) String() string {
	return fmt.Sprintf("local %s", c.Path)
}
