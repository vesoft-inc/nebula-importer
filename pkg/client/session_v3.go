//go:generate mockgen -source=session.go -destination session_mock.go -package client Session
package client

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type (
	defaultSessionV3 struct {
		session     *nebula.Session
		hostAddress nebula.HostAddress
		user        string
		password    string
		tlsConfig   *tls.Config
		logger      *slog.Logger
	}
)

func newSessionV3(hostAddress HostAddress, user, password string, tlsConfig *tls.Config, l *slog.Logger) Session {
	if l == nil {
		l = slog.Default()
	}
	return &defaultSessionV3{
		hostAddress: nebula.HostAddress{
			Host: hostAddress.Host,
			Port: hostAddress.Port,
		},
		user:      user,
		password:  password,
		tlsConfig: tlsConfig,
		logger:    l,
	}
}

func (s *defaultSessionV3) Open() error {
	hostAddress := s.hostAddress
	pool, err := nebula.NewSslConnectionPool(
		[]nebula.HostAddress{hostAddress},
		nebula.PoolConfig{
			MaxConnPoolSize: 1,
		},
		s.tlsConfig,
		newNebulaLogger(s.logger.With(
			"address",
			fmt.Sprintf("%s:%d", hostAddress.Host, hostAddress.Port),
		)),
	)
	if err != nil {
		return err
	}

	session, err := pool.GetSession(s.user, s.password)
	if err != nil {
		return err
	}

	s.session = session

	return nil
}

func (s *defaultSessionV3) Execute(statement string) (Response, error) {
	startTime := time.Now()
	rs, err := s.session.Execute(statement)
	if err != nil {
		return nil, err
	}
	return newResponseV3(rs, time.Since(startTime)), nil
}

func (s *defaultSessionV3) Close() error {
	s.session.Release()
	return nil
}
