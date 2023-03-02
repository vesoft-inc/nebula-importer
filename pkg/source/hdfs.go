package source

import (
	"fmt"
	"strings"

	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
)

var _ Source = (*hdfsSource)(nil)

type (
	HDFSConfig struct {
		Address string `yaml:"address,omitempty"`
		User    string `yaml:"user,omitempty"`
		Path    string `yaml:"path,omitempty"`
	}

	hdfsSource struct {
		c   *Config
		cli *hdfs.Client
		r   *hdfs.FileReader
	}
)

func newHDFSSource(c *Config) Source {
	return &hdfsSource{
		c: c,
	}
}

func (s *hdfsSource) Name() string {
	return s.c.HDFS.String()
}

func (s *hdfsSource) Open() error {
	// TODO: support kerberos
	conf, err := hadoopconf.LoadFromEnvironment()
	if err != nil {
		return err
	}

	options := hdfs.ClientOptionsFromConf(conf)
	if s.c.HDFS.Address != "" {
		options.Addresses = strings.Split(s.c.HDFS.Address, ",")
	}
	options.User = s.c.HDFS.User

	cli, err := hdfs.NewClient(options)
	if err != nil {
		return err
	}

	r, err := cli.Open(s.c.HDFS.Path)
	if err != nil {
		return err
	}

	s.cli = cli
	s.r = r

	return nil
}

func (s *hdfsSource) Config() *Config {
	return s.c
}

func (s *hdfsSource) Size() (int64, error) {
	return s.r.Stat().Size(), nil
}

func (s *hdfsSource) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *hdfsSource) Close() error {
	defer func() {
		_ = s.cli.Close()
	}()
	return s.r.Close()
}

func (c *HDFSConfig) String() string {
	return fmt.Sprintf("hdfs %s %s", c.Address, c.Path)
}
