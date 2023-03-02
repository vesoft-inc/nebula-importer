package source

import (
	"fmt"
	"time"

	"github.com/jlaffaye/ftp"
)

var _ Source = (*ftpSource)(nil)

type (
	FTPConfig struct {
		Host     string `yaml:"host,omitempty"`
		Port     int    `yaml:"port,omitempty"`
		User     string `yaml:"user,omitempty"`
		Password string `yaml:"password,omitempty"`
		Path     string `yaml:"path,omitempty"`
	}

	ftpSource struct {
		c    *Config
		conn *ftp.ServerConn
		r    *ftp.Response
		size int64
	}
)

func newFTPSource(c *Config) Source {
	return &ftpSource{
		c: c,
	}
}

func (s *ftpSource) Name() string {
	return s.c.FTP.String()
}

func (s *ftpSource) Open() error {
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", s.c.FTP.Host, s.c.FTP.Port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}

	err = conn.Login(s.c.FTP.User, s.c.FTP.Password)
	if err != nil {
		_ = conn.Quit()
		return err
	}

	size, err := conn.FileSize(s.c.FTP.Path)
	if err != nil {
		_ = conn.Quit()
		return err
	}

	r, err := conn.Retr(s.c.FTP.Path)
	if err != nil {
		_ = conn.Quit()
		return err
	}

	s.conn = conn
	s.r = r
	s.size = size

	return nil
}

func (s *ftpSource) Config() *Config {
	return s.c
}

func (s *ftpSource) Size() (int64, error) {
	return s.size, nil
}

func (s *ftpSource) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *ftpSource) Close() error {
	defer func() {
		_ = s.conn.Quit()
	}()
	return s.r.Close()
}

func (c *FTPConfig) String() string {
	return fmt.Sprintf("ftp %s:%d %s", c.Host, c.Port, c.Path)
}
